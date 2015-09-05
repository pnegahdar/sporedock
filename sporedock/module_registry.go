package sporedock
import (
	"github.com/pnegahdar/sporedock/types"
	"sync"
	"github.com/pnegahdar/sporedock/utils"
	"fmt"
	"time"
	"runtime/debug"
	"github.com/gorilla/mux"
	"net"
	"github.com/fsouza/go-dockerclient"
	"github.com/pnegahdar/sporedock/modules"
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

func (gr *ModuleRegistry) registerModules(modules ...types.Module) {
	gr.startMe = make(chan string, len(modules))
	// Todo: check should run
	utils.LogInfo(fmt.Sprintf("%v modules", len(modules)))
	for _, grunt := range modules {
		gruntName := grunt.ProcName()
		utils.LogInfo(fmt.Sprintf("Adding grunt %v", gruntName))
		gr.modules[gruntName] = grunt
		gr.runCount[gruntName] = 0
		gr.startMe <- gruntName
	}

}

func (gr *ModuleRegistry) runGrunt(gruntName string) {
	gr.Lock()
	grunt, exists := gr.modules[gruntName]
	if !exists {
		utils.LogWarn(fmt.Sprintf("Grunt %v DNE %v", gruntName, grunt))
		return
	}
	runCount := gr.runCount[gruntName]
	delayTot := RestartDecaySeconds * runCount
	gr.runCount[gruntName] = runCount + 1
	utils.LogInfo(fmt.Sprintf("Running grunt %v with delay of %v seconds", gruntName, delayTot))
	exit, _ := gr.stopCast.Listen()
	gr.Unlock()
	select {
	case <-time.After(time.Duration(delayTot) * time.Second):
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					utils.LogWarn(fmt.Sprintf("Grunt %v paniced", gruntName))
					debug.PrintStack()
				} else {
					utils.LogWarn(fmt.Sprintf("Grunt %v exited", gruntName))
				}
				gr.Lock()
				gr.startMe <- gruntName
				gr.Unlock()
			}()
			utils.LogInfo(fmt.Sprintf("Started grunt %v", gruntName))
			grunt.Run(gr.runContext)
		}()
	case <-exit:
		return
	}
}

func (gr *ModuleRegistry) Start(block bool, modules ...types.Module) {
	gr.registerModules(modules...)
	utils.LogInfo("Runner started.")
	for _, module := range modules {
		module.Init(gr.runContext)
	}
	runner := func() {
		exit, _ := gr.stopCast.Listen()
		for {
			select {
			case gruntToStart := <-gr.startMe:
				gr.runGrunt(gruntToStart)
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

func (gr *ModuleRegistry) Stop() {
	utils.LogInfo("Stopping grunt controller.")
	gr.stopCastMu.Lock()
	defer gr.stopCastMu.Unlock()
	gr.stopCast.Signal()

	for _, grunt := range gr.modules {
		grunt.Stop()
	}

}

func (gr *ModuleRegistry) Wait() {
	exit, _ := gr.stopCast.Listen()
	<-exit
}

func NewGruntRegistry(rc *types.RunContext) *ModuleRegistry {
	modules := make(map[string]types.Module)
	runCount := make(map[string]int)
	return &ModuleRegistry{runContext: rc, modules: modules, runCount: runCount}
}

func CreateAndRun(connectionString, groupName, machineID, machineIP string, webServerBind string, rpcServerBind string) *ModuleRegistry {
	myIP := net.ParseIP(machineIP)
	webServerRouter := mux.NewRouter().StrictSlash(true)

	// Create Run Context
	dockerClient, err := docker.NewClientFromEnv()
	utils.HandleError(err)
	runContext := types.RunContext{MyMachineID: machineID, MyIP: myIP, MyGroup: groupName, WebServerBind: webServerBind, WebServerRouter: webServerRouter, DockerClient: dockerClient, RPCServerBind: rpcServerBind}
	// Register and run
	gruntRegistry := NewGruntRegistry(&runContext)


	store := modules.CreateStore(&runContext, connectionString, groupName)
	api := &modules.SporeAPI{}
	webserver := &modules.WebServer{}
	planner := &modules.Planner{}
	dockerRunner := &modules.DockerRunner{}
	loadBalancer := &modules.LoadBalancer{}
	rpcserver := &modules.RPCServer{}
	runContext.Store = store

	gruntRegistry.Start(false, store, api, webserver, planner, dockerRunner, loadBalancer, rpcserver)
	return gruntRegistry
}
