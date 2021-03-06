package registry

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/pnegahdar/sporedock/modules"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

const RestartDecaySeconds = 1

type ModuleRegistry struct {
	sync.Mutex
	orderedModules []types.Module
	modules        map[string]types.Module
	runContext     *types.RunContext
	runCount       map[string]int
	startMe        chan string
	stopCast       utils.SignalCast
	stopCastMu     sync.Mutex
	initOnce       sync.Once
}

func (mr *ModuleRegistry) registerModules(modules ...types.Module) {
	mr.startMe = make(chan string, len(modules))
	// Todo: check should run
	utils.LogInfo(fmt.Sprintf("%v modules", len(modules)))
	mr.orderedModules = append(mr.orderedModules, modules...)
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

func (mr *ModuleRegistry) RegisterModules(modules ...types.Module) {
	utils.LogInfo("Registering modules")
	mr.registerModules(modules...)
}

func (mr *ModuleRegistry) RunModules(block bool, config *types.Config) {
	mr.runContext.Lock()
	mr.runContext.Config = config
	mr.runContext.Unlock()
	utils.LogInfo("Initializing Modules")
	for _, module := range mr.orderedModules {
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

func (mr *ModuleRegistry) RunCli() {
	for _, module := range mr.modules {
		if cliModule, ok := module.(types.CliModule); ok {
			cliModule.InitCli(mr.runContext)
		}
	}
	mr.runContext.CliManager.AddCommand(cli.Command{
		Name:      "run",
		ShortName: "r",
		Usage:     "Run sporedock main",
		Flags:     types.RunCommandFlags,
		Action: func(c *cli.Context) {
			mr.RunModules(true, types.NewConfigFromCli(c))
		}})
	mr.runContext.CliManager.Cli.Run(os.Args)
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

func Create() *ModuleRegistry {
	runContext := types.NewRunContext()
	// Register and run
	moduleRegistry := NewModuleRegistry(runContext)

	store := &modules.RedisStore{}
	api := &modules.SporeAPI{}
	webserver := &modules.WebServer{}
	eventServer := &modules.EventModule{}
	planner := &modules.PlannerModule{}
	dockerRunner := &modules.DockerRunnerModule{}
	loadBalancer := &modules.LoadBalancerModule{}
	rpcserver := &modules.RPCServer{}

	moduleRegistry.RegisterModules(store, webserver, api, eventServer, planner, dockerRunner, loadBalancer, rpcserver)
	return moduleRegistry
}
