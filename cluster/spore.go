package cluster

import (
	"errors"
	"github.com/pnegahdar/sporedock/types"
	"net"
)

var IPParseError = errors.New("The IP of the machine is not parsable as a standard IP.")

type Spore struct {
	ID         string
	MemberIP   string
	MemberType types.SporeType
	Tags       map[string]string
}

func (s *Spore) Validate() error {
	ok := net.ParseIP(s.MemberIP)
	if ok == nil {
		return IPParseError
	}
	return nil
}

func AllSpores(rc *types.RunContext) (*[]Spore, error) {
	sporeType := []Spore{}
	err := rc.Store.GetAll(&sporeType, 0, types.SentinelEnd)
	if err != nil {
		return nil, err
	}
	return &sporeType, err
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

func MySpore(rc *types.RunContext) (*Spore, error) {
	spore := &Spore{}
	err := rc.Store.Get(spore, rc.MyMachineID)
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
