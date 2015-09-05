package modules

import (
	"github.com/garyburd/redigo/redis"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"net"
	"strings"
	"sync"
	"time"
)

const CheckinEveryMs = 1000 //Delta between these two indicate how long it takes for something to be considered gone.
const CheckinExpireMs = 5000
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
	myType           types.SporeType
	myMachineID      string
	rc               *types.RunContext
	stopCast         utils.SignalCast
	stopCastMu       sync.Mutex
}

func (rs RedisStore) keyJoiner(runContext *types.RunContext, parts ...string) string {
	items := append(runContext.NamespacePrefixParts(), parts...)
	return strings.Join(items, ":")
}

func (rs RedisStore) typeKey(runContext *types.RunContext, v interface{}, parts ...string) string {
	meta, err := types.NewMeta(v)
	utils.HandleError(err)
	parts = append([]string{meta.TypeName}, parts...)
	return rs.keyJoiner(runContext, parts...)
}

func (rs RedisStore) runLeaderElection() {
	if rs.myType != types.TypeSporeWatcher {
		leaderKey := rs.keyJoiner(rs.rc, "_redis", "_leader")
		conn := rs.GetConn()
		defer conn.Close()
		_, err := conn.Do("SET", leaderKey, rs.myMachineID, "NX", "PX", LeadershipExpireMs)
		utils.HandleError(err)
	}
}

func (rs *RedisStore) runPruning() {
	spores := []cluster.Spore{}
	err := rs.GetAll(&spores, 0, types.SentinelEnd)
	utils.HandleError(err)
	for _, spore := range spores {
		healthy, err := rs.IsHealthy(spore.ID)
		utils.HandleError(err)
		if !healthy {
			utils.LogWarn("Spore" + spore.ID + "looks dead, purning.")
			err := rs.Delete(spore, spore.ID)
			utils.HandleError(err)
		}
	}

}

func (rs *RedisStore) runCheckIn() {
	conn := rs.GetConn()
	defer conn.Close()
	//Todo protect for duped names
	memberKey := rs.keyJoiner(rs.rc, "_redis", "_member", rs.myMachineID)
	leader, err := rs.LeaderName()
	utils.HandleError(err)
	if leader == rs.myMachineID {
		rs.mu.Lock()
		rs.myType = types.TypeSporeLeader
		rs.mu.Unlock()
	}
	spore := cluster.Spore{ID: rs.myMachineID, MemberIP: rs.myIP.String(), MemberType: rs.myType}
	err = rs.Update(spore, spore.ID, types.SentinelEnd)
	if err != types.ErrIDExists {
		utils.HandleError(err)
	}
	_, err = conn.Do("SET", memberKey, rs.myMachineID, "PX", CheckinExpireMs)
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

func (rs *RedisStore) Init(runContext *types.RunContext) {
	return
}

func (rs *RedisStore) Run(context *types.RunContext) {
	rs.mu.Lock()
	rs.stopCast = utils.SignalCast{}
	exit, _ := rs.stopCast.Listen()
	rs.mu.Unlock()
	rs.runLeaderElection()
	rs.runCheckIn()
	for {
		select {
		case <-time.After(time.Millisecond * CheckinEveryMs):
			rs.runLeaderElection()
			rs.runCheckIn()
			rs.runPruning()
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
	resp, err := conn.Do("HGET", rs.typeKey(rs.rc, v), id)
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
	resp, err := conn.Do("HEXISTS", rs.typeKey(rs.rc, v), id)
	exists, err := redis.Bool(resp, err)
	if err != nil {
		return false, wrapError(err)
	}
	return exists, nil

}

func (rs RedisStore) GetAll(v interface{}, start int, end int) error {
	conn := rs.GetConn()
	defer conn.Close()
	resp, err := conn.Do("HVALS", rs.typeKey(rs.rc, v))
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

func (rs RedisStore) safeSet(v interface{}, id string, logTrim int, update bool) error {
	conn := rs.GetConn()
	defer conn.Close()
	typeKey := rs.typeKey(rs.rc, v)
	data, err := utils.Marshall(v)
	if err != nil {
		return wrapError(err)
	}
	if id == "" {
		return types.ErrIDEmpty
	}
	op := "HSET"
	if !update {
		op = "HSETNX"
	}
	wasSet, err := redis.Int(conn.Do(op, typeKey, id, data))
	if err != nil {
		return wrapError(err)
	}
	if wasSet == 1 || update {
		logKey := rs.typeKey(rs.rc, v, "__log")
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

func (rs RedisStore) Set(v interface{}, id string, logTrim int) error {
	return rs.safeSet(v, id, logTrim, false)
}

func (rs RedisStore) Update(v interface{}, id string, logTrim int) error {
	return rs.safeSet(v, id, logTrim, true)
}

func (rs RedisStore) Delete(v interface{}, id string) error {
	conn := rs.GetConn()
	defer conn.Close()
	exists, err := redis.Int(conn.Do("HDEL", rs.typeKey(rs.rc, v), id))
	if exists != 1 {
		return types.ErrNoneFound
	}
	return wrapError(err)
}

func (rs RedisStore) DeleteAll(v interface{}) error {
	conn := rs.GetConn()
	defer conn.Close()
	_, err := conn.Do("DEL", rs.typeKey(rs.rc, v))
	return wrapError(err)
}

func (rs RedisStore) IsHealthy(sporeName string) (bool, error) {
	conn := rs.GetConn()
	defer conn.Close()
	memberKey := rs.keyJoiner(rs.rc, "_redis", "_member", sporeName)
	resp, err := conn.Do("EXISTS", memberKey)
	exists, err := redis.Bool(resp, err)
	if err != nil {
		return false, wrapError(err)
	}
	return exists, nil
}

func (rs RedisStore) LeaderName() (string, error) {
	leaderKey := rs.keyJoiner(rs.rc, "_redis", "_leader")
	conn := rs.GetConn()
	defer conn.Close()
	name, err := redis.String(conn.Do("GET", leaderKey))
	return name, wrapError(err)
}

func NewRedisStore(context *types.RunContext, redisConnecitonURI, group string) types.SporeStore {
	redisConnecitonURI = strings.TrimPrefix(redisConnecitonURI, "redis://")
	return &RedisStore{rc: context, connectionString: redisConnecitonURI, group: group}
}
