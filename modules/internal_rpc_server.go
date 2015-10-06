package modules

import (
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"github.com/valyala/gorpc"
	"sync"
)

type RPCServer struct {
	initOnce   sync.Once
	stopCast   utils.SignalCast
	runContext *types.RunContext
	server     *gorpc.Server
}

func (rpc *RPCServer) Init(runContext *types.RunContext) {
	rpc.initOnce.Do(func() {
		// Requires one fake func to run/init
		rpcManager := (&types.RPCManager{RPCServerBind: runContext.Config.RPCServerBind}).Init()
		rpcManager.RPCAddFunc("_interfal_fake", func() {})
		runContext.Lock()
		runContext.RPCManager = rpcManager
		runContext.Unlock()
		rpc.runContext = runContext
		rpc.server = gorpc.NewTCPServer(rpcManager.RPCServerBind, rpcManager.RPCDispatcher().NewHandlerFunc())
	})
}

func (rpc *RPCServer) ShouldRun(runContext *types.RunContext) bool {
	return true
}

func (rpc *RPCServer) ProcName() string {
	return "RPCServer"
}

func (rpc *RPCServer) Stop() {
	rpc.runContext.RPCManager.RPCCloseAll()
	rpc.server.Stop()
	rpc.stopCast.Signal()
}

func (rpc *RPCServer) Run(runContext *types.RunContext) {
	exit, _ := rpc.stopCast.Listen()
	rpc.server.Start()
	<-exit
}
