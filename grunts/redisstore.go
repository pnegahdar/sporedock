package grunts

import (
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
const LeadershipExpireMs = 3000

func CreateStore(context *types.RunContext, connectionString, group string) types.SporeStore {
	if strings.HasPrefix(connectionString, "redis://") {
		return NewRedisStore(context, connectionString, group)
	} else {
		utils.HandleError(types.ErrConnectionString)
		return nil
	}
}

var knownErrors = map[error]error{redis.ErrNil: types.ErrNoneFound}

func wrapError(err error) error {
	rewrap, ok := knownErrors[err]
	if ok {
		return rewrap
	}
	return err
}

type RedisStore struct {
	mu               sync.Mutex
	ConnMu           sync.Mutex
	initOnce         sync.Once
	connectionString string
	connPool         *redis.Pool
	group            string
	myIP             net.IP
	amLeader         bool
	myType           types.SporeType
	myMachineID      string
	rc               *types.RunContext
	stopCast         utils.SignalCast
	stopCastMu       sync.Mutex
}

func (rs RedisStore) keyJoiner(parts ...string) string {
	items := []string{"sporedock", rs.group}
	for _, part := range parts {
		items = append(items, part)
	}
	return strings.Join(items, ":")
}

func (rs RedisStore) typeKey(v interface{}, parts ...string) string {
	meta, err := types.NewMeta(v)
	utils.HandleError(err)
	parts = append([]string{meta.TypeName}, parts...)
	return rs.keyJoiner(parts...)
}

func (rs RedisStore) runLeaderElection() {
	leaderKey := rs.keyJoiner("_redis", "_leader")
	conn := rs.GetConn()
	_, err := conn.Do("SET", leaderKey, rs.myMachineID, "NX", "PX", LeadershipExpireMs)
	utils.HandleError(err)
	conn.Close()
	// Todo: what if this fails
}

func (rs *RedisStore) runCheckIn() {
	conn := rs.GetConn()
	defer conn.Close()
	_, err := conn.Do("PSETEX", rs.myMachineID, CheckinExpireMs, "1")
	utils.HandleError(err)
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
	rs.mu.Lock()
	rs.stopCast = utils.SignalCast{}
	exit, _ := rs.stopCast.Listen()
	rs.mu.Unlock()
	for {
		select {
		case <-time.After(time.Millisecond * CheckinEveryMs):
			rs.runCheckIn()
			rs.runLeaderElection()
		case <-exit:
			return
		}
	}
}

func (rs *RedisStore) setup() {
	if rs.connectionString == "" {
		utils.HandleError(types.ErrConnectionStringNotSet)
	}
	rs.group = rs.rc.MyGroup
	rs.myIP = rs.rc.MyIP
	rs.myType = rs.rc.MyType
	rs.myMachineID = rs.rc.MyMachineID
	rs.connPool = newRedisConnPool(rs.connectionString)
}

func (rs *RedisStore) GetConn() redis.Conn {
	rs.ConnMu.Lock()
	defer rs.ConnMu.Unlock()
	rs.initOnce.Do(rs.setup)
	conn := rs.connPool.Get()
	return conn
}

func (rs *RedisStore) Stop() {
	rs.stopCastMu.Lock()
	defer rs.stopCastMu.Unlock()
	rs.stopCast.Signal()

}

func (rs RedisStore) ProcName() string {
	return "RedisStore"
}

func (rs RedisStore) ShouldRun(context *types.RunContext) bool {
	return true
}

func (rs RedisStore) Get(v interface{}, id string) error {
	conn := rs.GetConn()
	defer conn.Close()
	resp, err := conn.Do("HGET", rs.typeKey(v), id)
	data, err := redis.String(resp, err)
	if err != nil {
		return wrapError(err)
	}
	err = utils.Unmarshall(data, v)
	if err != nil {
		return wrapError(err)
	}
	return nil
}

func (rs RedisStore) Exists(v interface{}, id string) (bool, error) {
	conn := rs.GetConn()
	defer conn.Close()
	resp, err := conn.Do("HEXISTS", rs.typeKey(v), id)
	exists, err := redis.Bool(resp, err)
	if err != nil {
		return false, wrapError(err)
	}
	return exists, nil

}

func (rs RedisStore) GetAll(v interface{}, start int, end int) error {
	conn := rs.GetConn()
	defer conn.Close()
	resp, err := conn.Do("HVALS", rs.typeKey(v))
	data, err := redis.Strings(resp, err)
	if err != nil {
		return wrapError(err)
	}
	if end == types.SentinelEnd {
		end = len(data)
	}
	joined := utils.JsonListFromObjects(data[start:end]...)
	err = utils.Unmarshall(joined, v)
	if err != nil {
		return err
	}
	return nil
}

func (rs RedisStore) Set(v interface{}, id string, logTrim int) error {
	typeKey := rs.typeKey(v)
	data, err := utils.Marshall(v)
	if err != nil {
		return wrapError(err)
	}
	if id == "" {
		return types.ErrIDEmpty
	}
	conn := rs.GetConn()
	defer conn.Close()
	wasSet, err := redis.Int(conn.Do("HSETNX", typeKey, id, data))
	if err != nil {
		return wrapError(err)
	}
	if wasSet == 1 {
		logKey := rs.typeKey(v, "__log")
		_, err = conn.Do("LPUSH", logKey, data)
		if err != nil {
			return wrapError(err)
		}
		if logTrim != types.SentinelEnd {
			_, err = conn.Do("LTRIM", logKey, 0, logTrim)
			if err != nil {
				return wrapError(err)
			}
		}
		return nil
	} else {
		return types.ErrIDExists
	}
}

func (rs RedisStore) Delete(v interface{}, id string) error {
	conn := rs.GetConn()
	defer conn.Close()
	exists, err := redis.Int(conn.Do("HDEL", rs.typeKey(v), id))
	if exists != 1 {
		return types.ErrNoneFound
	}
	return wrapError(err)
}

func (rs RedisStore) DeleteAll(v interface{}) error {
	conn := rs.GetConn()
	defer conn.Close()
	_, err := conn.Do("DEL", rs.typeKey(v))
	return wrapError(err)
}

func NewRedisStore(context *types.RunContext, redisConnecitonURI, group string) types.SporeStore {
	redisConnecitonURI = strings.TrimPrefix(redisConnecitonURI, "redis://")
	return &RedisStore{rc: context, connectionString: redisConnecitonURI, group: group}
}
