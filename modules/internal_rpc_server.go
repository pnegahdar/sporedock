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
	// Requires one fake func to run/init
	runContext.RPCAddFunc("_interfal_fake", func(){})
	rpc.initOnce.Do(func() {
		rpc.runContext = runContext
		rpc.server = gorpc.NewTCPServer(runContext.RPCServerBind, runContext.RPCDispatcher().NewHandlerFunc())
	})
}

func (rpc *RPCServer) ShouldRun(runContext *types.RunContext) bool {
	return true
}

func (rpc *RPCServer) ProcName() string {
	return "RPCServer"
}

func (rpc *RPCServer) Stop() {
	rpc.runContext.RPCCloseAll()
	rpc.server.Stop()
	rpc.stopCast.Signal()
}

func (rpc *RPCServer) Run(runContext *types.RunContext) {
	exit, _ := rpc.stopCast.Listen()
	rpc.server.Start()
	<-exit
}
