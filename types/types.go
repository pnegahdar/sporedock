package types

import (
	"net"
)

type SporeType string

const (
	TypeSporeLeader  SporeType = "leader"
	TypeSporeMember  SporeType = "member"
	TypeSporeWatcher SporeType = "watcher"
)

type SporeStore interface {
	ProcName() string
	ShouldRun(RunContext) bool
	Get(item Storable) Storable
	GetAll(retType Storable) []Storable
	GetLog(retType Storable, limit int) []Storable
	Set(item Storable)
	SetLog(item Storable, logLength int)
	Delete(item Storable)
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
	FromString(data string) (Storable, error)
}
