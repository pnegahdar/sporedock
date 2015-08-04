package types

import (
	"errors"
	"fmt"
	"net"
	"reflect"
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

var ErrConnectionString = errors.New("Connection string must start with redis://")
var ErrConnectionStringNotSet = errors.New("Connection string not set.")

// HTTP status errors
type Grunt interface {
	ProcName() string
	Run(runContext *RunContext)
	Stop()
	ShouldRun(runContext *RunContext) bool
}

const SentinelEnd = -1

type SporeStore interface {
	Grunt
	Get(i interface{}, id string) error
	GetAll(v interface{}, start int, end int) error
	Set(v interface{}, id string, logTrim int) error
	Exists(v interface{}, id string) (bool, error)
	Delete(v interface{}, id string) error
	DeleteAll(v interface{}) error
}

type RunContext struct {
	Store       SporeStore
	MyMachineID string
	MyIP        net.IP
	MyType      SporeType
	MyGroup     string
}

type TypeMeta struct {
	IsStruct bool
	TypeName string
}

func NewMeta(v interface{}) (TypeMeta, error) {
	typeof := reflect.TypeOf(v)
	kind := typeof.Kind()
	fmt.Println(kind)
	fmt.Println(typeof)
	if kind == reflect.Ptr {
		typeof = reflect.ValueOf(v).Elem().Type()
		kind = typeof.Kind()
	}
	fmt.Println(kind)
	fmt.Println(typeof)

	switch kind {
	case reflect.Slice:
		meta := TypeMeta{IsStruct: true, TypeName: fmt.Sprint(typeof.Elem())}
		return meta, nil
	case reflect.Struct:
		meta := TypeMeta{IsStruct: false, TypeName: fmt.Sprint(typeof)}
		return meta, nil
	default:
		err := errors.New("Type not struct or slice")
		return TypeMeta{}, err

	}
}
