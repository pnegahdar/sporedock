package discovery

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
	MemberIP   net.IP
	MemberType SporeType
}

type Serializable interface {
	SerialKey() string
	Serialize() string
	Deserialize(data string) (*Serializable, error)
}

// Todo (parham): Key locking with with some sort of watch interface
type SporeStore interface {
	ListMembers() []Spore
	GetLeader() Spore
	GetMe() Spore
	AmLeader() bool
	GetKey(key string) (string, error)
	SetKey(key, value string) error
	SetKeyWithLog(key, value string, logLength int) error
	Load(load_into *Serializable) (*Serializable, error)
	Save(to_save Serializable) error
	Run(group string, myType SporeType, myIP net.IP)
}

func GetStore() SporeStore {
	// Return redis store
	if CurrentStore != nil {
		return CurrentStore, nil
	}
	if strings.HasPrefix(connectionString, "redis://") {
		return &RedisStore{connectionString}
	} else {
		utils.HandleError(ConnectionStringError)
		return nil
	}
}
