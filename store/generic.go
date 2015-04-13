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

var CurrentStore SporeStore

type Storable interface {
	TypeIdentifier() string
	Identifier() string
	ToString() string
	FromString(data string) (Storable, error)
}

type SporeStore interface {
	MyID() string
    Get(item Storable) Storable
    GetAll(retType Storable) []Storable
    GetLog(retType Storable, limit int) []Storable
    Set(item Storable)
    SetLog(item Storable, logLength int) error
	Delete(item Storable)
	Run(group string, myType cluster.SporeType, myIP net.IP)
}

func CreateStore(connectionString, group string) SporeStore {
    if (CurrentStore != SporeStore{}){
        return CurrentStore
    }
	if strings.HasPrefix(connectionString, "redis://") {
		CurrentStore = NewRedisStore(connectionString, group)
        return CurrentStore
	} else {
		utils.HandleError(ConnectionStringError)
		return nil
	}
}
