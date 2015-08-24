package types

import (
	"errors"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
	"net"
	"reflect"
	"strings"
	"net/rpc"
	"sync"
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

type Identifiable interface {
	GetID() string
}

type Creatable interface {
	Identifiable
	Validate(*RunContext) error
}

type Validable interface {
	Identifiable
	Validate(*RunContext) error
}

const CpuMemMultiplier = 1 / 512

type Sizable struct {
	Cpus float64
	Mem  float64
}

func GetSize(cpu float64, mem float64) float64 {
	return cpu + (mem * CpuMemMultiplier)
}

type SporeStore interface {
	Grunt
	Get(i interface{}, id string) error
	GetAll(v interface{}, start int, end int) error
	Set(v interface{}, id string, logTrim int) error
	Update(v interface{}, id string, logTrim int) error
	Exists(v interface{}, id string) (bool, error)
	Delete(v interface{}, id string) error
	DeleteAll(v interface{}) error
	IsHealthy(sporeName string) (bool, error)
	LeaderName() (string, error)
}

type RunContext struct {
	Store           SporeStore
	MyMachineID     string
	MyIP            net.IP
	MyType          SporeType
	MyGroup         string
	WebServerBind   string
	RPCServerBind   string
	WebServerRouter *mux.Router
	DockerClient    *docker.Client
	rpcOnce         sync.Once
	RPCServer       *rpc.Server
}

func (rc RunContext) NamespacePrefixParts() []string {
	return []string{"sporedock", rc.MyGroup, rc.MyMachineID}
}

func (rc RunContext) NamespacePrefix(joiner string, extra ...string) string {
	data := append(rc.NamespacePrefixParts(), extra...)
	return strings.Join(data, joiner)
}

func (rc *RunContext) initRPC() {
	rc.rpcOnce.Do(func() {
		rc.RPCServer = rpc.NewServer()
	})
}

func (rc *RunContext) RPCRegister(rcvr interface{}) {
	rc.initRPC()
	rc.RPCServer.Register(rcvr)
}


type TypeMeta struct {
	IsSlice  bool
	TypeName string
}

func NewMeta(v interface{}) (TypeMeta, error) {
	typeof := reflect.TypeOf(v)
	kind := typeof.Kind()
	if kind == reflect.Ptr {
		typeof = reflect.ValueOf(v).Elem().Type()
		kind = typeof.Kind()
	}
	var isSlice bool
	var typeName string
	switch kind {
	case reflect.Slice:
		isSlice = true
		typeName = fmt.Sprint(typeof.Elem())
	case reflect.Struct:
		typeName = fmt.Sprint(typeof)
		isSlice = false
	case reflect.Interface:
		isSlice = false
		typeName = fmt.Sprint(reflect.ValueOf(v).Elem().Elem().Type())
	case reflect.String, reflect.Int, reflect.Bool:
		isSlice = false
		typeName = fmt.Sprint(typeof)
	default:
		err := errors.New("Type not struct or slice")
		return TypeMeta{}, err
	}
	return TypeMeta{IsSlice: isSlice, TypeName: strings.TrimPrefix(typeName, "*")}, nil
}
