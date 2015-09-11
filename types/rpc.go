package types

import (
	"github.com/valyala/gorpc"
	"sync"
)

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
