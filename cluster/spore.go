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
