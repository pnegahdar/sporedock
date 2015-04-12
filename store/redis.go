package store

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
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
}

const CheckinEveryMs = 1000 //Delta between these two indicate how long it takes for something to be considered gone.
const CheckinExpireMs = 3000

const LeadershipCheckinMs = 3000
const LeadershipExpireMs = 5000

var OneExitedError = errors.New("One of hte proceses exited")
var BadKeyError = errors.New("The key provided wasn't in the the right format")
var BadIPError = errors.New("The IP failed to parse")

func (rs RedisStore) groupKey(storable Storable) string {
	return strings.Join([]string{"sporedock", rs.group, storable.TypeIdentifier()}, ":")
}

func (rs RedisStore) itemKey(storable Storable) string {
	return strings.Join([]string{rs.groupKey(storable), storable.Identifier()}, ":")
}

func (rs RedisStore) subItemKey(storable Storable, subitems ...string) {
	items := append([]string{rs.itemKey(storable)}, subitems)
	return strings.Join(items, ":")
}

func (rs RedisStore) leaderKey() string {
	return strings.Join([]string{"sporedock", rs.group, "_leader"}, ":")
}

func (rs RedisStore) myKey() string {
	return strings.Join()
	return fmt.Sprintf("sporedock:members:%v:%v:%v", rs.group, rs.myIP.String(), rs.myType)
}

func (rs RedisStore) membersKey() string {
	return fmt.Sprintf("sporedock:members:*", rs.group)
}

func (rs RedisStore) getMachineFromKey(key string) (Spore, error) {
	data := strings.Split(key, ":")
	if len(data) != 5 {
		return nil, BadKeyError
	}
	group := data[2]
	memberIP := net.ParseIP(data[3])
	if memberIP == nil {
		return nil, BadIPError
	}
	// Todo will this enum casting work?
	memberType := SporeType(data[4])
	return Spore{Group: group, MemberIP: memberIP, SporeType: memberType}, nil

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
	rs.group = group
	rs.myIP = myIP
	rs.myType = myType
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

func (rs RedisStore) ListMembers() []Spore {
	reply, err := rs.connection.Do("KEYS", rs.membersKey())
	utils.HandleError(err)
	var members []Spore
	resp, err := redis.Strings(reply, nil)
	utils.HandleError(err)
	for _, key := range resp {
		newMember, err := rs.getMachineFromKey(key)
		utils.HandleError(err)
		members = append(members, newMember)
	}
	return members
}

func (rs RedisStore) GetLeader() Spore {
	reply, err := rs.connection.Do("GET", rs.leaderKey())
	utils.HandleError(err)
	resp, err := redis.String(reply, nil)
	utils.HandleError(err)
	member, err := rs.getMachineFromKey(resp)
	utils.HandleError(err)
	return member
}

func (rs RedisStore) GetMe() Spore {
	reply, err := rs.connection.Do("GET", rs.myKey())
	utils.HandleError(err)
	resp, err := redis.String(reply, nil)
	utils.HandleError(err)
	machine, err := rs.getMachineFromKey(resp)
	utils.HandleError(err)
	return machine
}

func (rs RedisStore) AmLeader() bool {
	leader := rs.GetLeader()
	if leader.MemberIP == rs.myIP {
		return true
	}
	return false
}

func (rs RedisStore) GetKey(key string) (string, error) {
	resp, err := rs.connection.Do("GET", key)
	if err != nil {
		return nil, err
	} else {

		return redis.String(resp, nil)
	}
}

func (rs RedisStore) SetKey(key, value string) error {
	_, err := rs.connection.Do("SET", key, value)
	if err != nil {
		return err
	}
	return nil
}

func (rs RedisStore) SetKeyWithLog(key, value string, logLength int) {
	err := rs.SetKey(key, value)
	utils.HandleError(err)
	logKey := fmt.Sprintf("%v__log", key)
	_, err = rs.connection.Do("LPUSH", logKey, value)
	utils.HandleError(err)
	_, err = rs.connection.Do("LTRIM", key, 0, logLength)
	utils.HandleError(err)
}

func (rs RedisStore) Save(to_save Storable) error {
	err := rs.SetKey(to_save.SerialKey(), to_save.Serialize())
	return err
}

func (rs RedisStore) Load(load_into Storable) (*Storable, error) {
	data, err := rs.GetKey(load_into.SerialKey())
	if err != nil {
		return nil, error
	}
	ret_val, err := load_into.Deserialize(data)
	return ret_val, error
}

func NewRedisStore(redisConnecitonURI, group string) RedisStore {
	return &RedisStore{connectionString: redisConnecitonURI, group: group}

}
