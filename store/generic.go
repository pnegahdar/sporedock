package store

import (
	"errors"
	"github.com/pnegahdar/sporedock/utils"
	"net"
	"strings"
)

type SporeType int

var CurrentStore SporeStore

const (
	TypeSporeLeader SporeType = iota
	TypeSporeMember
	TypeSporeWatcher
)

var ConnectionStringError = errors.New("Connection string must start with redis://")
var ConnectionStringNotSetError = errors.New("Connection string not set.")

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

func (s Spore) MyIdentifier() string {
	return s.Name
}

func (s Spore) ToString() string {
	return utils.Marshall(s)
}

func (s Spore) validate() error {
	return nil
}

func (s *Spore) FromString(data string) (*Storable, error) {
	utils.Unmarshall(data, s)
	err := s.validate()
	return s, err
}

type Storable interface {
	TypeIdentifier() string
	MyIdentifer() string
	ToString() string
	FromString(data string) (Storable, error)
}

// Todo (parham): Key locking with with some sort of watch interface
type SporeStore interface {
	ListMembers() []Spore
	GetLeader() Spore
	GetMe() Spore
	GetGroupName() string
	AmLeader() bool
	GetAll(Storable) []Storable
	Set(item Storable)
	SetLog(item Storable, logLength int) error
	Get(item Storable) Storable
	GetLog(item Storable, limit int) []Storable
	Delete(item Storable)
	Run(group string, myType SporeType, myIP net.IP)
}

func GetStore(connectionString string, group string) SporeStore {
	// Return redis store
	if CurrentStore != nil {
		return CurrentStore
	}
	if strings.HasPrefix(connectionString, "redis://") {
		return &RedisStore{connectionString}
	} else {
		utils.HandleError(ConnectionStringError)
		return nil
	}
}
