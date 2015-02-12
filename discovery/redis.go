package discovery

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

const OneExitedError = errors.New("One of hte proceses exited")
const BadKeyError = errors.New("The key provided wasn't in the the right format")
const BadIPError = errors.New("The IP failed to parse")

func (rs RedisStore) lockKey() string {
	return fmt.Sprintf("sporedock:leader:%v", rs.group)
}

func (rs RedisStore) myKey() string {
	return fmt.Sprintf("sporedock:members:%v:%v:%v", rs.group, rs.myIP.String(), rs.myType)
}

func (rs RedisStore) membersKey() string {
	return fmt.Sprintf("sporedock:members:*", rs.group)
}

func (rs RedisStore) getMachineFromKey(key string) (Member, error) {
	data := strings.Split(key, ":")
	if len(data) != 5 {
		return nil, BadKeyError
	}
	group := data[2]
	memberIP := net.ParseIP(data[3])
	if memberIP == nil {
		return nil, BadIPError
	}
	memberType := MemberType(data[4])
	return Member{Group: group, MemberIP: memberIP, MemberType: memberType}, nil

}

func (rs RedisStore) runLeaderElection(wg sync.WaitGroup) {
	lockKey := rs.lockKey()
	myKey := rs.myKey()
	resp, err := rs.connection.Do("SETNX", lockKey, myKey)
	utils.HandleErrorWG(err, wg)
	if resp.(int64) == 1 {
		_, err := rs.connection.Do("PEXPIRE", lockKey, LeadershipExpireMs)
		utils.HandleErrorWG(err, wg)
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

func (rs RedisStore) Run(group string, myType MemberType, myIP net.IP) {
	if rs.connectionString == nil {
		utils.HandleError(ConnectionStringNotSetError)
	}
	connection, err := redis.Dial("tcp", rs.connectionString)
	utils.HandleError(err)
	rs.group = group
	rs.myIP = myIP
	rs.myType = myType
	rs.connection = connection
	// Todo(parham): better proc management here
	// All or none WG Group
	wg := sync.WaitGroup{}
	wg.Add(1)
	go rs.runCheckIn(wg)
	go rs.runLeaderElection(wg)
	wg.Wait()
	utils.HandleError(OneExitedError)
}

func (rs RedisStore) ListMembers() []Member {
	resp, err := rs.connection.Do("KEYS", rs.membersKey())
	utils.HandleError(err)
	var members []Member
	for _, key := range resp.([]string) {
		newMember, err := rs.getMachineFromKey(key)
		utils.HandleError(err)
		members = append(members, newMember)
	}
	return members
}

func (rs RedisStore) GetLeader() Member {
	resp, err := rs.connection.Do("GET", rs.lockKey())
	utils.HandleError(err)
	member, err := rs.getMachineFromKey(resp.(string))
	utils.HandleError(err)
	return member
}

func (rs RedisStore) GetMe() Member {
	resp, err := rs.connection.Do("GET", rs.myKey())
	utils.HandleError(err)
	machine, err := rs.getMachineFromKey(resp.(string))
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

func (rs RedisStore) SetKey(key, value string) error {
	_, err := rs.connection.Do("GET", key)
	if err != nil {
		return err
	}
	return nil
}

func (rs RedisStore) GetKey(key string) (string, error) {
	resp, err := rs.connection.Do("GET", key)
	if err != nil {
		return nil, err
	} else {
		return (resp).(string), err
	}
}
