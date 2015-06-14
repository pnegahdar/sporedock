package grunts

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"net"
	"strings"
	"sync"
	"time"
)

const CheckinEveryMs = 1000 //Delta between these two indicate how long it takes for something to be considered gone.
const CheckinExpireMs = 3000

const LeadershipCheckinMs = 3000
const LeadershipExpireMs = 5000

// Which errors to pipe through to next layer (say web), which to panic on.
var remapErrors = map[error]types.HttpError{redis.ErrNil: types.ErrEmptyQuery}

var CurrentStore types.SporeStore

func CreateStore(connectionString, group string) types.SporeStore {
	if CurrentStore != nil {
		return CurrentStore
	}
	if strings.HasPrefix(connectionString, "redis://") {
		CurrentStore = NewRedisStore(connectionString, group)
		return CurrentStore
	} else {
		utils.HandleError(types.ErrConnectionString)
		return nil
	}
}

type RedisStore struct {
	connectionString string
	connPool         *redis.Pool
	group            string
	myIP             net.IP
	amLeader         bool
	myType           types.SporeType
	myMachineID      string
	rc               *types.RunContext
}

var OneExitedError = errors.New("One of the proceses exited")

func (rs RedisStore) typeKey(storable types.Storable) string {
	return strings.Join([]string{"sporedock", rs.group, storable.TypeIdentifier()}, ":")
}

func (rs RedisStore) itemKey(storable types.Storable) string {
	return strings.Join([]string{rs.typeKey(storable), storable.Identifier(), "*"}, ":")
}

func (rs RedisStore) subItemKey(storable types.Storable, subitems ...string) string {
	items := []string{rs.itemKey(storable)}
	for _, subitem := range subitems {
		items = append(items, subitem)
	}
	return strings.Join(items, ":")
}

func (rs RedisStore) leaderKey() string {
	return strings.Join([]string{"sporedock", rs.group, "_redis", "_leader"}, ":")
}

func (rs RedisStore) logKey(storable types.Storable) string {
	return fmt.Sprint("%v__log", rs.itemKey(storable))
}

func (rs RedisStore) membersKey() string {
	return fmt.Sprintf("sporedock:members:*", rs.group)
}

func (rs RedisStore) runLeaderElection(wg sync.WaitGroup) {
	checkinDur := time.Millisecond * LeadershipCheckinMs
	leaderKey := rs.leaderKey()
	for {
		conn := rs.connPool.Get()
		reply, err := conn.Do("SETNX", leaderKey, rs.myMachineID)
		utils.HandleErrorWG(err, wg)
		resp, err := redis.Int(reply, nil)
		utils.HandleError(err)
		if resp == 1 {
			_, err := conn.Do("PEXPIRE", leaderKey, LeadershipExpireMs)
			utils.HandleErrorWG(err, wg)
		}
		conn.Close()
		time.Sleep(checkinDur)
	}

}

func (rs RedisStore) runCheckIn(wg sync.WaitGroup) {
	// TODO FIX
	checkinDur := time.Millisecond * CheckinEveryMs
	for {
		conn := rs.connPool.Get()
		_, err := conn.Do("PSETEX", rs.myMachineID, CheckinExpireMs, "1")
		utils.HandleErrorWG(err, wg)
		time.Sleep(checkinDur)
		conn.Close()
	}

}

func newRedisConnPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (rs *RedisStore) Run(context *types.RunContext) {
	if rs.connectionString == "" {
		utils.HandleError(types.ErrConnectionStringNotSet)
	}
	// Todo: Connection pool
	rs.group = context.MyGroup
	rs.myIP = context.MyIP
	rs.myType = context.MyType
	rs.myMachineID = context.MyMachineID
	rs.connPool = newRedisConnPool(rs.connectionString)
	rs.rc = context
	// Todo: better proc management here
	// All or none WG Group
	wg := sync.WaitGroup{}
	wg.Add(1)
	go rs.runCheckIn(wg)
	go rs.runLeaderElection(wg)
	wg.Wait()
	utils.HandleError(OneExitedError)
}

func (rs RedisStore) Get(retType types.Storable) (types.Storable, error) {
	conn := rs.connPool.Get()
	defer conn.Close()
	resp, err := conn.Do("GET", rs.itemKey(retType))
	data, err := redis.String(resp, err)
	if err != nil {
		return nil, err
	}
	obj, err := retType.FromString(data, rs.rc)
	if err != nil {
		return nil, err
	}
	return obj, nil

}

func (rs RedisStore) ProcName() string {
	return "RedisStore"
}

func (rs RedisStore) GetAll(retType types.Storable) ([]types.Storable, error) {
	// Todo: Switch to hash table as this is slow
	conn := rs.connPool.Get()
	defer conn.Close()
	resp, err := conn.Do("KEYS", rs.typeKey(retType))
	keys, err := redis.Strings(resp, err)
	if err != nil {
		return nil, err
	}

	conn.Send("MULTI")
	for _, key := range keys {
		conn.Send("GET", key)
	}
	resp, err = conn.Do("EXEC")
	data, err := redis.Strings(resp, err)
	if err != nil {
		return nil, err
	}
	storables := []types.Storable{}
	for _, storableString := range data {
		obj, err := retType.FromString(storableString, rs.rc)
		if err != nil {
			utils.LogWarn(fmt.Sprintf("Was unable to parse from DB item %v. Please delete or fix. Skipping for now. Body: %v", rs.typeKey(retType), storableString))
			continue
		}
		storables = append(storables, obj)
	}
	return storables, nil
}

func (rs RedisStore) GetLog(retType types.Storable, limit int) ([]types.Storable, error) {
	conn := rs.connPool.Get()
	defer conn.Close()
	resp, err := conn.Do("LRANGE", rs.typeKey(retType), 0, limit)
	data, err := redis.Strings(resp, err)
	if err != nil {
		return nil, err
	}
	storables := []types.Storable{}
	for _, storableString := range data {
		obj, err := retType.FromString(storableString, rs.rc)
		if err != nil {
			utils.LogWarn(fmt.Sprintf("Was unable to parse from DB item %v. Please delete or fix. Skipping for now. Body: %v", rs.typeKey(retType), storableString))
			continue
		}
		storables = append(storables, obj)
	}
	return storables, nil
}

func (rs RedisStore) Set(item types.Storable) error {
	key := rs.itemKey(item)
	data := item.ToString()
	conn := rs.connPool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, data)
	return err
}

func (rs RedisStore) SetLog(item types.Storable, logLength int) error {
	err := rs.Set(item)
	if err != nil {
		return err
	}
	logKey := fmt.Sprint("%v__log", rs.itemKey(item))
	conn := rs.connPool.Get()
	defer conn.Close()
	_, err = conn.Do("LPUSH", logKey, item.ToString())
	if err != nil {
		return err
	}
	_, err = conn.Do("LTRIM", logKey, 0, logLength)
	return err
}

func (rs RedisStore) Delete(item types.Storable) error {
	conn := rs.connPool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", rs.itemKey(item))
	return err
}

func (rs RedisStore) ShouldRun(context types.RunContext) bool {
	return true
}

func NewRedisStore(redisConnecitonURI, group string) types.SporeStore {
	redisConnecitonURI = strings.TrimPrefix(redisConnecitonURI, "redis://")
	return &RedisStore{connectionString: redisConnecitonURI, group: group}
}
