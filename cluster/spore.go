package cluster

import (
	"errors"
	"github.com/pnegahdar/sporedock/utils"
	"net"
)

type SporeType string

var IPParseError = errors.New("The IP of the machine is not parsable as a standard IP.")

const (
	TypeSporeLeader  SporeType = "leader"
	TypeSporeMember  SporeType = "member"
	TypeSporeWatcher SporeType = "watcher"
)

type Spore struct {
	Group      string
	Name       string
	MemberIP   string
	MemberType SporeType
	Tags       map[string]string
}

func (s Spore) TypeIdentifier() string {
	return "spore"
}

func (s Spore) Identifier() string {
	return s.Name
}

func (s Spore) ToString() string {
	data, err := utils.Marshall(s)
	utils.HandleError(err)
	return data
}

func (s Spore) validate() error {
	ok := net.ParseIP(s.MemberIP)
	if ok == nil {
		return IPParseError
	}
	return nil
}

func (s Spore) FromString(data string) (Spore, error) {
	s = Spore{}
	utils.Unmarshall(data, &s)
	err := s.validate()
	return s, err
}

//func Members(rc grunts.RunContext) []Spore {
//	sporeType := Spore{}
//	return store.CurrentStore.GetAll(sporeType).([]Spore)
//}
//
//func Leader() Spore {
//	// Todo: Don't scan all
//	members := Members("YO")
//	for _, member := range members {
//		if member.MemberType == TypeSporeLeader {
//			return member
//		}
//	}
//	return nil
//}
//
//func Me() Spore {
//	members := Members()
//	for _, member := range members {
//		if member.Name == store.CurrentStore.MyID() {
//			return member
//		}
//
//	}
//	return nil
//}
//
//func AmLeader() bool {
//	leader := Leader()
//	if leader.Name == store.CurrentStore.MyID() {
//		return true
//	}
//	return false
//}
