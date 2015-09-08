package cluster

import (
	"errors"
	"github.com/pnegahdar/sporedock/types"
	"net"
	"net/rpc"
)

var IPParseError = errors.New("The IP of the machine is not parsable as a standard IP.")

type SporeID string

type Spore struct {
	ID         string
	MemberIP   string
	MemberType types.SporeType
	Tags       map[string]string
	types.Sizable
}

func (spore *Spore) RPCCall(serviceMethod string, args interface{}, reply interface{}) error {
	// Todo bypass if local
	// Todo unfix ip
	client, err := rpc.DialHTTP("tcp", spore.MemberIP+":5001")
	if err != nil {
		return err
	}
	defer client.Close()
	err = client.Call(serviceMethod, args, reply)
	return err
}
func (s Spore) Size() float64 {
	return types.GetSize(s.Cpus, s.Mem)
}

type Spores []Spore

func (s Spores) Len() int           { return len(s) }
func (s Spores) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Spores) Less(i, j int) bool { return s[i].Size() < s[j].Size() }

func (s *Spore) Validate() error {
	ok := net.ParseIP(s.MemberIP)
	if ok == nil {
		return IPParseError
	}
	return nil
}

func AllSpores(rc *types.RunContext) (Spores, error) {
	sporeType := []Spore{}
	err := rc.Store.GetAll(&sporeType, 0, types.SentinelEnd)
	if err != nil {
		return nil, err
	}
	return Spores(sporeType), err
}

func LeaderSpore(rc *types.RunContext) (*Spore, error) {
	spore := &Spore{}
	leaderName, err := rc.Store.LeaderName()
	if err != nil {
		return nil, err
	}
	err = rc.Store.Get(spore, leaderName)
	if err != nil {
		return nil, err
	}
	return spore, nil
}

func GetSpore(rc *types.RunContext, id string) (*Spore, error) {
	spore := &Spore{}
	err := rc.Store.Get(spore, id)
	if err != nil {
		return nil, err
	}
	return spore, nil
}

func AmLeader(rc *types.RunContext) (bool, error) {
	leaderName, err := rc.Store.LeaderName()
	if err != nil {
		return false, err
	}
	return rc.MyMachineID == leaderName, nil

}
