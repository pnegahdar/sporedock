package store

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/pnegahdar/sporedock/grunts"
	"github.com/pnegahdar/sporedock/utils"
	"net"
	"strings"
	"sync"
	"time"
)

type RedisStore struct {
	connectionString string
	connection       redis.Conn
	group            string
	myIP             net.IP
	amLeader         bool
	myType           int
	myMachineID      string
}

const CheckinEveryMs = 1000 //Delta between these two indicate how long it takes for something to be considered gone.
const CheckinExpireMs = 3000

const LeadershipCheckinMs = 3000
const LeadershipExpireMs = 5000

var OneExitedError = errors.New("One of hte proceses exited")

func (rs RedisStore) typeKey(storable Storable) string {
	return strings.Join([]string{"sporedock", rs.group, storable.TypeIdentifier()}, ":")
}

func (rs RedisStore) itemKey(storable Storable) string {
	return strings.Join([]string{rs.typeKey(storable), storable.Identifier(), "*"}, ":")
}

func (rs RedisStore) subItemKey(storable Storable, subitems ...string) {
	items := append([]string{rs.itemKey(storable)}, subitems)
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
	myKey := rs.myKey()
	for {
		reply, err := rs.connection.Do("SETNX", leaderKey, myKey)
		utils.HandleErrorWG(err, wg)
		resp, err := redis.Int(reply, nil)
		utils.HandleError(err)
		if resp == 1 {
			_, err := rs.connection.Do("PEXPIRE", leaderKey, LeadershipExpireMs)
			utils.HandleErrorWG(err, wg)
		}
		time.Sleep(checkinDur)
	}

}

func (rs RedisStore) runCheckIn(wg sync.WaitGroup) {
	checkinDur := time.Millisecond * CheckinEveryMs
	for {
		_, err := rs.connection.Do("PSETEX", rs.myKey(), CheckinExpireMs, "1")
		utils.HandleErrorWG(err, wg)
		time.Sleep(checkinDur)
	}

}

func (rs RedisStore) Run(context grunts.RunContext) {
	if rs.connectionString == nil {
		utils.HandleError(ConnectionStringNotSetError)
	}
	// Todo: Connection pool
	connection, err := redis.Dial("tcp", rs.connectionString)
	utils.HandleError(err)
	rs.group = context.myGroup
	rs.myIP = context.myIP
	rs.myType = context.myType
	rs.myMachineID = context.myMachineID
	rs.connection = connection
	// Todo: better proc management here
	// All or none WG Group
	wg := sync.WaitGroup{}
	wg.Add(1)
	go rs.runCheckIn(wg)
	go rs.runLeaderElection(wg)
	wg.Wait()
	utils.HandleError(OneExitedError)
}

func (rs RedisStore) Get(retType Storable) {
	resp, err := rs.connection.Do("GET", rs.itemKey(retType))
	data, err := redis.String(resp, err)
	utils.HandleError(err)
	return retType.FromString(data)

}

func (rs RedisStore) GetAll(retType Storable) []Storable {
	// Todo: Switch to hash table as this is slow
	resp, err := rs.connection.Do("KEYS", rs.typeKey(retType))
	keys, err := redis.Strings(resp, err)
	utils.HandleError(err)

	rs.connection.Send("MULTI")
	for _, key := range keys {
		rs.connection.Send("GET", key)
	}
	resp, err = rs.connection.Do("EXEC")
	data, err := redis.Strings(resp, err)
	utils.HandleError(err)

	storables := []Storable{}
	for _, storable := range data {
		storables = append(storables, retType.FromString(storable))
	}
	return storables
}

func (rs RedisStore) GetLog(retType Storable, limit int) {
	resp, err := rs.connection.Do("LRANGE", rs.typeKey(retType), 0, limit)
	data, err := redis.Strings(resp, err)
	utils.HandleError(err)
	storables := []Storable{}
	for _, item := range data {
		append(storables, retType.FromString(item))
	}
}

func (rs RedisStore) Set(item Storable) {
	key := rs.itemKey(item)
	data := item.ToString()
	_, err := rs.connection.Do("SET", key, data)
	utils.HandleError(err)
	return nil
}

func (rs RedisStore) SetWithLog(item Storable, logLength int) {
	rs.Set(item)
	logKey := fmt.Sprint("%v__log", rs.itemKey(item))
	_, err := rs.connection.Do("LPUSH", logKey, item.ToString())
	utils.HandleError(err)
	_, err = rs.connection.Do("LTRIM", logKey, 0, logLength)
	utils.HandleError(err)
}

func (rs RedisStore) Delete(item Storable) {
	_, err := rs.connection.Do("DEL", rs.itemKey(item))
	utils.HandleError(err)

}

func NewRedisStore(redisConnecitonURI, group string) RedisStore {
	return &RedisStore{connectionString: redisConnecitonURI, group: group}

}
