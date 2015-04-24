package grunts

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/pnegahdar/sporedock/cluster"
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

var ConnectionStringError = errors.New("Connection string must start with redis://")
var ConnectionStringNotSetError = errors.New("Connection string not set.")

var CurrentStore SporeStore

type Storable interface {
	TypeIdentifier() string
	Identifier() string
	ToString() string
	FromString(data string) (Storable, error)
}

type SporeStore interface {
	ProcName() string
	ShouldRun(RunContext) bool
	Get(item Storable) Storable
	GetAll(retType Storable) []Storable
	GetLog(retType Storable, limit int) []Storable
	Set(item Storable)
	SetLog(item Storable, logLength int)
	Delete(item Storable)
	Run(context *RunContext)
}

func CreateStore(connectionString, group string) SporeStore {
	if CurrentStore != nil {
		return CurrentStore
	}
	if strings.HasPrefix(connectionString, "redis://") {
		CurrentStore = NewRedisStore(connectionString, group)
		return CurrentStore
	} else {
		utils.HandleError(ConnectionStringError)
		return nil
	}
}

type RedisStore struct {
	connectionString string
	connPool         *redis.Pool
	group            string
	myIP             net.IP
	amLeader         bool
	myType           cluster.SporeType
	myMachineID      string
}

var OneExitedError = errors.New("One of hte proceses exited")

func (rs RedisStore) typeKey(storable Storable) string {
	return strings.Join([]string{"sporedock", rs.group, storable.TypeIdentifier()}, ":")
}

func (rs RedisStore) itemKey(storable Storable) string {
	return strings.Join([]string{rs.typeKey(storable), storable.Identifier(), "*"}, ":")
}

func (rs RedisStore) subItemKey(storable Storable, subitems ...string) string {
	items := []string{rs.itemKey(storable)}
	for _, subitem := range subitems {
		items = append(items, subitem)
	}
	return strings.Join(items, ":")
}

func (rs RedisStore) leaderKey() string {
	return strings.Join([]string{"sporedock", rs.group, "_redis", "_leader"}, ":")
}

func (rs RedisStore) logKey(storable Storable) string {
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

func (rs RedisStore) Run(context *RunContext) {
	if rs.connectionString == "" {
		utils.HandleError(ConnectionStringNotSetError)
	}
	// Todo: Connection pool
	rs.group = context.myGroup
	rs.myIP = context.myIP
	rs.myType = context.myType
	rs.myMachineID = context.myMachineID
	rs.connPool = newRedisConnPool(rs.connectionString)
	// Todo: better proc management here
	// All or none WG Group
	wg := sync.WaitGroup{}
	wg.Add(1)
	go rs.runCheckIn(wg)
	go rs.runLeaderElection(wg)
	wg.Wait()
	utils.HandleError(OneExitedError)
}

func (rs RedisStore) Get(retType Storable) Storable {
	conn := rs.connPool.Get()
	defer conn.Close()
	resp, err := conn.Do("GET", rs.itemKey(retType))
	data, err := redis.String(resp, err)
	utils.HandleError(err)
	obj, err := retType.FromString(data)
	utils.HandleError(err)
	return obj

}

func (rs RedisStore) ProcName() string {
	return "RedisStore"
}

func (rs RedisStore) GetAll(retType Storable) []Storable {
	// Todo: Switch to hash table as this is slow
	conn := rs.connPool.Get()
	defer conn.Close()
	resp, err := conn.Do("KEYS", rs.typeKey(retType))
	keys, err := redis.Strings(resp, err)
	utils.HandleError(err)

	conn.Send("MULTI")
	for _, key := range keys {
		conn.Send("GET", key)
	}
	resp, err = conn.Do("EXEC")
	data, err := redis.Strings(resp, err)
	utils.HandleError(err)

	storables := []Storable{}
	for _, storable := range data {
		obj, err := retType.FromString(storable)
		utils.HandleError(err)
		storables = append(storables, obj)
	}
	return storables
}

func (rs RedisStore) GetLog(retType Storable, limit int) []Storable {
	conn := rs.connPool.Get()
	defer conn.Close()
	resp, err := conn.Do("LRANGE", rs.typeKey(retType), 0, limit)
	data, err := redis.Strings(resp, err)
	utils.HandleError(err)
	storables := []Storable{}
	for _, item := range data {
		obj, err := retType.FromString(item)
		utils.HandleError(err)
		storables = append(storables, obj)
	}
	return storables
}

func (rs RedisStore) Set(item Storable) {
	key := rs.itemKey(item)
	data := item.ToString()
	conn := rs.connPool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, data)
	utils.HandleError(err)
}

func (rs RedisStore) SetLog(item Storable, logLength int) {
	rs.Set(item)
	logKey := fmt.Sprint("%v__log", rs.itemKey(item))
	conn := rs.connPool.Get()
	defer conn.Close()
	_, err := conn.Do("LPUSH", logKey, item.ToString())
	utils.HandleError(err)
	_, err = conn.Do("LTRIM", logKey, 0, logLength)
	utils.HandleError(err)
}

func (rs RedisStore) Delete(item Storable) {
	conn := rs.connPool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", rs.itemKey(item))
	utils.HandleError(err)

}

func (rs RedisStore) ShouldRun(context RunContext) bool {
	return true
}

func NewRedisStore(redisConnecitonURI, group string) RedisStore {
	redisConnecitonURI = strings.TrimPrefix(redisConnecitonURI, "redis://")
	return RedisStore{connectionString: redisConnecitonURI, group: group}
}
