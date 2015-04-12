package store

import (
	"errors"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/utils"
	"net"
	"strings"
)

var ConnectionStringError = errors.New("Connection string must start with redis://")
var ConnectionStringNotSetError = errors.New("Connection string not set.")

type Storable interface {
	TypeIdentifier() string
	Identifier() string
	ToString() string
	FromString(data string) (Storable, error)
}

type SporeStore interface {
	Members() []cluster.Spore
	Leader() cluster.Spore
	Me() cluster.Spore
	GroupName() string
	AmLeader() bool
	GetAll(retType Storable) []Storable
	Set(item Storable)
	SetLog(item Storable, logLength int) error
	Get(item Storable) Storable
	GetLog(item Storable, limit int) []Storable
	Delete(item Storable)
	Run(group string, myType cluster.SporeType, myIP net.IP)
}

func CreateStore(connectionString, group string) SporeStore {
	if strings.HasPrefix(connectionString, "redis://") {
		return NewRedisStore(connectionString, group)
	} else {
		utils.HandleError(ConnectionStringError)
		return nil
	}
}
