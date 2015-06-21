package types

import (
	"errors"
	"net"
)

type SporeType string

const (
	TypeSporeLeader  SporeType = "leader"
	TypeSporeMember  SporeType = "member"
	TypeSporeWatcher SporeType = "watcher"
)

type HttpError struct {
	Status int
	Error  error
}

func RewrapError(err interface{}) HttpError {
	switch err.(type) {
	case HttpError:
		return err.(HttpError)
	case error:
		// Return all other errors as a 400 bad request
		return HttpError{Status: 400, Error: err.(error)}
	default:
		panic("Unknown error type returned")
	}
}

var ErrConnectionString = errors.New("Connection string must start with redis://")
var ErrConnectionStringNotSet = errors.New("Connection string not set.")

// HTTP status errors
var ErrEmptyQuery = HttpError{Status: 404, Error: errors.New("No results found.")}

type SporeStore interface {
	ProcName() string
	ShouldRun(RunContext) bool
	Get(item Storable) (Storable, error)
	GetAll(retType Storable) ([]Storable, error)
	GetLog(retType Storable, limit int) ([]Storable, error)
	Set(item Storable) error
	SetLog(item Storable, logLength int) error
	Delete(item Storable) error
	Run(context *RunContext)
}

type RunContext struct {
	Store       SporeStore
	MyMachineID string
	MyIP        net.IP
	MyType      SporeType
	MyGroup     string
}

type Storable interface {
	TypeIdentifier() string
	Identifier() string
	ToString() string
	FromString(data string, rc *RunContext) (Storable, error)
}