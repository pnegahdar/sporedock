package sporedock

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
	"github.com/pnegahdar/sporedock/modules"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"net"
	"runtime/debug"
	"sync"
	"time"
)

const RestartDecaySeconds = 1

type ModuleRegistry struct {
	sync.Mutex
	modules    map[string]types.Module
	runContext *types.RunContext
	runCount   map[string]int
	startMe    chan string
	stopCast   utils.SignalCast
	stopCastMu sync.Mutex
}

func (mr *ModuleRegistry) registerModules(modules ...types.Module) {
	mr.startMe = make(chan string, len(modules))
	// Todo: check should run
	utils.LogInfo(fmt.Sprintf("%v modules", len(modules)))
	for _, module := range modules {
		moduleName := module.ProcName()
		utils.LogInfo(fmt.Sprintf("Adding module %v", moduleName))
		mr.modules[moduleName] = module
		mr.runCount[moduleName] = 0
		mr.startMe <- moduleName
	}

}

func (mr *ModuleRegistry) runModule(moduleName string) {
	mr.Lock()
	module, exists := mr.modules[moduleName]
	if !exists {
		utils.LogWarn(fmt.Sprintf("Module %v DNE %v", moduleName, module))
		return
	}
	runCount := mr.runCount[moduleName]
	delayTot := RestartDecaySeconds * runCount
	mr.runCount[moduleName] = runCount + 1
	utils.LogInfo(fmt.Sprintf("Running module %v with delay of %v seconds", moduleName, delayTot))
	exit, _ := mr.stopCast.Listen()
	mr.Unlock()
	select {
	case <-time.After(time.Duration(delayTot) * time.Second):
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					utils.LogWarn(fmt.Sprintf("Module %v paniced", moduleName))
					debug.PrintStack()
				} else {
					utils.LogWarn(fmt.Sprintf("Module %v exited", moduleName))
				}
				mr.Lock()
				mr.startMe <- moduleName
				mr.Unlock()
			}()
			utils.LogInfo(fmt.Sprintf("Started module %v", moduleName))
			module.Run(mr.runContext)
		}()
	case <-exit:
		return
	}
}

func (mr *ModuleRegistry) Start(block bool, modules ...types.Module) {
	mr.registerModules(modules...)
	utils.LogInfo("Runner started.")
	for _, module := range modules {
		module.Init(mr.runContext)
	}
	runner := func() {
		exit, _ := mr.stopCast.Listen()
		for {
			select {
			case moduleToStart := <-mr.startMe:
				mr.runModule(moduleToStart)
			case <-exit:
				return
			}
		}
	}
	if block {
		runner()
	} else {
		go runner()
	}
}

func (mr *ModuleRegistry) Stop() {
	utils.LogInfo("Stopping module controller.")
	mr.stopCastMu.Lock()
	defer mr.stopCastMu.Unlock()
	mr.stopCast.Signal()

	for _, module := range mr.modules {
		module.Stop()
	}

}

func (mr *ModuleRegistry) Wait() {
	exit, _ := mr.stopCast.Listen()
	<-exit
}

func NewModuleRegistry(rc *types.RunContext) *ModuleRegistry {
	return &ModuleRegistry{runContext: rc, modules: map[string]types.Module{}, runCount: map[string]int{}}
}

func CreateAndRun(connectionString, groupName, machineID, machineIP string, webServerBind string, rpcServerBind string) *ModuleRegistry {
	myIP := net.ParseIP(machineIP)
	webServerRouter := mux.NewRouter().StrictSlash(true)

	// Create Run Context
	dockerClient, err := docker.NewClientFromEnv()
	utils.HandleError(err)

	// Setup Managers
	rpcManager := (&types.RPCManager{RPCServerBind: rpcServerBind}).Init()
	webserverManager := &types.WebServerManager{WebServerBind: webServerBind, WebServerRouter: webServerRouter}

	runContext := types.RunContext{MyMachineID: machineID, MyIP: myIP, MyGroup: groupName, WebServerManager: webserverManager, DockerClient: dockerClient, RPCManager: rpcManager}
	// Register and run
	moduleRegistry := NewModuleRegistry(&runContext)

	store := modules.CreateStore(&runContext, connectionString, groupName)
	runContext.Store = store
	api := &modules.SporeAPI{}
	webserver := &modules.WebServer{}
	eventServer := &modules.EventModule{}
	planner := &modules.Planner{}
	dockerRunner := &modules.DockerRunner{}
	loadBalancer := &modules.LoadBalancer{}
	rpcserver := &modules.RPCServer{}

	moduleRegistry.Start(false, store, api, webserver, eventServer, planner, dockerRunner, loadBalancer, rpcserver)
	return moduleRegistry
}
