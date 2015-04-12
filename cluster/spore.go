package cluster

import "github.com/pnegahdar/sporedock/utils"

type SporeType string

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
	return utils.Marshall(s)
}

func (s Spore) validate() error {
	return nil
}

func (s *Spore) FromString(data string) (*Spore, error) {
	utils.Unmarshall(data, s)
	err := s.validate()
	return s, err
}
