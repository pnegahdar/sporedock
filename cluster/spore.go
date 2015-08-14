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
	Cpus       int
	Mem        int
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

func AllSporesMap(runContext *types.RunContext) (map[string]*Spore, error) {
	sporeMap := map[string]*Spore{}
	spores, err := AllSpores(runContext)
	if err != nil {
		return sporeMap, err
	}
	for _, spore := range spores {
		sporeMap[spore.ID] = &spore
	}
	return sporeMap, nil
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
