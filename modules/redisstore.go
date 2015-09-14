package modules

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
const CheckinExpireMs = 5000
const LeadershipExpireMs = 3000
const PubSubChannelNamePrefix = "pubsub"

func CreateStore(connectionString, group string) types.SporeStore {
	if strings.HasPrefix(connectionString, "redis://") {
		return NewRedisStore(connectionString)
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
	initOnce         sync.Once
	connectionString string
	connPool         *redis.Pool
	group            string
	myIP             net.IP
	myType           types.SporeType
	myMachineID      string
	runContext       *types.RunContext
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
		leaderKey := rs.keyJoiner(rs.runContext, "_redis", "_leader")
		conn := rs.GetConn()
		defer conn.Close()
		leaderChange, err := redis.String(conn.Do("SET", leaderKey, rs.myMachineID, "NX", "PX", LeadershipExpireMs))
		if err != redis.ErrNil {
			utils.HandleError(err)
		}
		if leaderChange == "OK" {
			types.EventStoreLeaderChange.EmitAll(rs.runContext)
		}
		leader, err := rs.LeaderName()
		utils.HandleError(err)
		if leader == rs.runContext.MyMachineID {
			_, err = conn.Do("PEXPIRE", leaderKey, LeadershipExpireMs)
			utils.HandleError(err)
		}
	}
}

func (rs *RedisStore) runPruning() {
	spores := []types.Spore{}
	err := rs.GetAll(&spores, 0, types.SentinelEnd)
	utils.HandleError(err)
	for _, spore := range spores {
		healthy, err := rs.IsHealthy(spore.ID)
		utils.HandleError(err)
		if !healthy {
			utils.LogWarn("Spore" + spore.ID + "looks dead, purning.")
			err := rs.Delete(spore, spore.ID)
			utils.HandleError(err)
			types.EventStoreSporeExit.EmitAll(rs.runContext)
		}
	}

}

func (rs *RedisStore) runCheckIn() {
	conn := rs.GetConn()
	defer conn.Close()
	//Todo protect for duped names
	memberKey := rs.keyJoiner(rs.runContext, "_redis", "_member", rs.myMachineID)
	leader, err := rs.LeaderName()
	utils.HandleError(err)
	if leader == rs.myMachineID {
		rs.mu.Lock()
		rs.myType = types.TypeSporeLeader
		rs.mu.Unlock()
	}
	spore := types.Spore{ID: rs.myMachineID, MemberIP: rs.myIP.String(), MemberType: rs.myType}
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
	rs.group = rs.runContext.MyGroup
	rs.myIP = rs.runContext.MyIP
	rs.myType = rs.runContext.MyType
	rs.myMachineID = rs.runContext.MyMachineID
	rs.connPool = newRedisConnPool(rs.connectionString)
}

func (rs *RedisStore) Init(runContext *types.RunContext) {
	rs.initOnce.Do(func() {
		rs.runContext = runContext
		rs.setup()
		runContext.Lock()
		runContext.Store = rs
		runContext.Unlock()
	})
	return
}

func (rs *RedisStore) GetConn() redis.Conn {
	conn := rs.connPool.Get()
	if conn.Err() != nil {
		utils.HandleError(conn.Err())
	}
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
	resp, err := conn.Do("HGET", rs.typeKey(rs.runContext, v), id)
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
	resp, err := conn.Do("HEXISTS", rs.typeKey(rs.runContext, v), id)
	exists, err := redis.Bool(resp, err)
	if err != nil {
		return false, wrapError(err)
	}
	return exists, nil

}

func (rs RedisStore) GetAll(v interface{}, start int, end int) error {
	conn := rs.GetConn()
	defer conn.Close()
	resp, err := conn.Do("HVALS", rs.typeKey(rs.runContext, v))
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
	typeKey := rs.typeKey(rs.runContext, v)
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
		meta, err := types.NewMeta(v)
		utils.HandleError(err)
		action := types.StoreActionUpdate
		if !update {
			action = types.StoreActionCreate
		}
		types.StoreEvent(action, meta).EmitAll(rs.runContext)
		logKey := rs.typeKey(rs.runContext, v, "__log")
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
	exists, err := redis.Int(conn.Do("HDEL", rs.typeKey(rs.runContext, v), id))
	if exists != 1 {
		return types.ErrNoneFound
	}
	if err != nil {
		return wrapError(err)
	}
	meta, err := types.NewMeta(v)
	utils.HandleError(err)
	types.StoreEvent(types.StoreActionDelete, meta).EmitAll(rs.runContext)
	return nil

}

func (rs RedisStore) DeleteAll(v interface{}) error {
	conn := rs.GetConn()
	defer conn.Close()
	_, err := conn.Do("DEL", rs.typeKey(rs.runContext, v))
	if err != nil {
		return wrapError(err)
	}
	meta, err := types.NewMeta(v)
	utils.HandleError(err)
	types.StoreEvent(types.StoreActionDeleteAll, meta).EmitAll(rs.runContext)
	return nil
}

func (rs RedisStore) IsHealthy(sporeName string) (bool, error) {
	conn := rs.GetConn()
	defer conn.Close()
	memberKey := rs.keyJoiner(rs.runContext, "_redis", "_member", sporeName)
	resp, err := conn.Do("EXISTS", memberKey)
	exists, err := redis.Bool(resp, err)
	if err != nil {
		return false, wrapError(err)
	}
	return exists, nil
}

func (rs RedisStore) LeaderName() (string, error) {
	leaderKey := rs.keyJoiner(rs.runContext, "_redis", "_leader")
	conn := rs.GetConn()
	defer conn.Close()
	name, err := redis.String(conn.Do("GET", leaderKey))
	return name, wrapError(err)
}

func (rs RedisStore) Publish(v interface{}, channels ...string) error {
	dump, err := utils.Marshall(v)
	if err != nil {
		return err
	}
	conn := rs.GetConn()
	defer conn.Close()
	for _, channel := range channels {
		fullChanName := rs.keyJoiner(rs.runContext, PubSubChannelNamePrefix, channel)
		conn.Send("PUBLISH", fullChanName, dump)
	}
	conn.Flush()
	_, err = conn.Receive()
	return err
}

func (rs RedisStore) Subscribe(channel string) (*types.SubscriptionManager, error) {
	messages := make(chan string)
	sm := &types.SubscriptionManager{ID: utils.GenGuid(), Messages: messages, Exit: utils.SignalCast{}}
	conn := rs.GetConn()
	psc := redis.PubSubConn{conn}
	fullChanName := rs.keyJoiner(rs.runContext, PubSubChannelNamePrefix, channel)
	err := psc.Subscribe(fullChanName)
	if err != nil {
		return nil, err
	}
	go func() {
		defer psc.Close()
		data := make(chan interface{})
		go func() {
			exit, _ := sm.Exit.Listen()
			for {
				select {
				case <-time.Tick(time.Millisecond * 200):
					dat := psc.Receive()
						select {
						case data <- dat:
						default:
							return
						}
				case <-exit:
					return
				}
			}
		}()
		exit, _ := sm.Exit.Listen()
		for {
			select {
			case <-exit:
				return
			case dat := <-data:
				switch v := dat.(type) {
				case redis.Message:
					go func() { select {
						case sm.Messages <- string(v.Data):
						default:
							return
						}
					}()
				case redis.Subscription:
					continue
				case error:
					utils.HandleError(v)
				}
			}
		}
	}()
	return sm, nil
}

func NewRedisStore(redisConnecitonURI string) types.SporeStore {
	redisConnecitonURI = strings.TrimPrefix(redisConnecitonURI, "redis://")
	return &RedisStore{connectionString: redisConnecitonURI}
}
