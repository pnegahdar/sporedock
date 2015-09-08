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

type RPCManager struct {
	RPCServerBind string
	RPCServer     *gorpc.Server
	RPCClients    map[string]*gorpc.Client
	rpcDispatcher *gorpc.Dispatcher
	clientLock    sync.Mutex
	initOnce      sync.Once
	sync.Mutex
}

func (rpcm *RPCManager) Init() *RPCManager {
	rpcm.initOnce.Do(func() {
		rpcm.rpcDispatcher = gorpc.NewDispatcher()
		rpcm.RPCClients = map[string]*gorpc.Client{}
	})
	return rpcm
}

func (rpcm *RPCManager) RPCAddFunc(funcName string, f interface{}) {
	rpcm.Lock()
	defer rpcm.Unlock()
	rpcm.rpcDispatcher.AddFunc(funcName, f)
}

func (rpcm *RPCManager) RPCDispatcher() *gorpc.Dispatcher {
	return rpcm.rpcDispatcher
}

func (rpcm *RPCManager) RPCCall(addr string, funcName string, request interface{}) (interface{}, error) {
	var rpcClient *gorpc.Client
	rpcm.clientLock.Lock()
	if rpcClient, ok := rpcm.RPCClients[addr]; !ok {
		rpcClient = gorpc.NewTCPClient(addr)
		rpcClient.Start()
		rpcm.RPCClients[addr] = rpcClient
	}
	rpcm.clientLock.Unlock()
	funcClient := rpcm.rpcDispatcher.NewFuncClient(rpcClient)
	return funcClient.Call(funcName, request)
}

func (rpcm *RPCManager) RPCCloseAll() {
	rpcm.Lock()
	defer rpcm.Unlock()
	for key, client := range rpcm.RPCClients {
		client.Stop()
		delete(rpcm.RPCClients, key)
	}
}

type WebServerManager struct {
	WebServerBind   string
	WebServerRouter *mux.Router
	sync.Mutex
}

type RunContext struct {
	Store            SporeStore
	MyMachineID      string
	MyIP             net.IP
	MyType           SporeType
	MyGroup          string
	RPCManager       *RPCManager
	WebServerManager *WebServerManager
	DockerClient     *docker.Client
	initOnce         sync.Once
	sync.Mutex
}

func (rc RunContext) NamespacePrefixParts() []string {
	return []string{"sporedock", rc.MyGroup, rc.MyMachineID}
}

func (rc RunContext) NamespacePrefix(joiner string, extra ...string) string {
	data := append(rc.NamespacePrefixParts(), extra...)
	return strings.Join(data, joiner)
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
