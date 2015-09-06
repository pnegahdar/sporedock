package types

import (
	"errors"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
	"github.com/valyala/gorpc"
	"net"
	"reflect"
	"strings"
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
type Module interface {
	ProcName() string
	ShouldRun(runContext *RunContext) bool
	Init(runContext *RunContext)
	Run(runContext *RunContext)
	Stop()
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
	Module
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
	RPCServer       *gorpc.Server
	RPCClients      map[string]*gorpc.Client
	rpcDispatcher   *gorpc.Dispatcher
	clientLock      sync.Mutex
	DockerClient    *docker.Client
	initOnce        sync.Once
	rpcAddLock      sync.Mutex
}

func (rc RunContext) NamespacePrefixParts() []string {
	return []string{"sporedock", rc.MyGroup, rc.MyMachineID}
}

func (rc RunContext) NamespacePrefix(joiner string, extra ...string) string {
	data := append(rc.NamespacePrefixParts(), extra...)
	return strings.Join(data, joiner)
}

func (rc *RunContext) Init() {
	rc.initOnce.Do(func() {
		rc.rpcDispatcher = gorpc.NewDispatcher()
		rc.RPCClients = map[string]*gorpc.Client{}
	})
}

func (rc *RunContext) RPCAddFunc(funcName string, f interface{}) {
	rc.Init()
	rc.rpcAddLock.Lock()
	defer rc.rpcAddLock.Unlock()
	rc.rpcDispatcher.AddFunc(funcName, f)
}

func (rc *RunContext) RPCDispatcher() *gorpc.Dispatcher {
	rc.Init()
	return rc.rpcDispatcher
}

func (rc *RunContext) RPCCall(addr string, funcName string, request interface{}) (interface{}, error) {
	rc.Init()
	var rpcClient *gorpc.Client
	rc.clientLock.Lock()
	if rpcClient, ok := rc.RPCClients[addr]; !ok {
		rpcClient = gorpc.NewTCPClient(addr)
		rpcClient.Start()
		rc.RPCClients[addr] = rpcClient
	}
	rc.clientLock.Unlock()
	funcClient := rc.rpcDispatcher.NewFuncClient(rpcClient)
	return funcClient.Call(funcName, request)
}

func (rc *RunContext) RPCCloseAll() {
	rc.rpcAddLock.Lock()
	defer rc.rpcAddLock.Unlock()
	for key, client := range rc.RPCClients {
		client.Stop()
		delete(rc.RPCClients, key)
	}
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
