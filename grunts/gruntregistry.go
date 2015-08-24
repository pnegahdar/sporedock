package grunts

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"net"
	"runtime/debug"
	"sync"
	"time"
)

const RestartDecaySeconds = 1

type GruntRegistry struct {
	sync.Mutex
	Grunts     map[string]types.Grunt
	Context    *types.RunContext
	runCount   map[string]int
	startMe    chan string
	stopCast   utils.SignalCast
	stopCastMu sync.Mutex
}

func (gr *GruntRegistry) registerGrunts(grunts ...types.Grunt) {
	gr.startMe = make(chan string, len(grunts))
	// Todo: check should run
	utils.LogInfo(fmt.Sprintf("%v grunts", len(grunts)))
	for _, grunt := range grunts {
		gruntName := grunt.ProcName()
		utils.LogInfo(fmt.Sprintf("Adding grunt %v", gruntName))
		gr.Grunts[gruntName] = grunt
		gr.runCount[gruntName] = 0
		gr.startMe <- gruntName
	}

}

func (gr *GruntRegistry) runGrunt(gruntName string) {
	gr.Lock()
	grunt, exists := gr.Grunts[gruntName]
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
			grunt.Run(gr.Context)
		}()
	case <-exit:
		return
	}
}

func (gr *GruntRegistry) Start(grunts ...types.Grunt) {
	gr.registerGrunts(grunts...)
	utils.LogInfo("Runner started.")
	// Range blocks on startMe channel
	go func() {
		exit, _ := gr.stopCast.Listen()
		for {
			select {
			case gruntToStart := <-gr.startMe:
				gr.runGrunt(gruntToStart)
			case <-exit:
				return
			}
		}
	}()
}

func (gr *GruntRegistry) Stop() {
	utils.LogInfo("Stopping grunt controller.")
	gr.stopCastMu.Lock()
	defer gr.stopCastMu.Unlock()
	gr.stopCast.Signal()

	for _, grunt := range gr.Grunts {
		grunt.Stop()
	}

}

func (gr *GruntRegistry) Wait() {
	exit, _ := gr.stopCast.Listen()
	<-exit
}

func NewGruntRegistry(rc *types.RunContext) *GruntRegistry {
	grunts := make(map[string]types.Grunt)
	runCount := make(map[string]int)
	return &GruntRegistry{Context: rc, Grunts: grunts, runCount: runCount}
}

func CreateAndRun(connectionString, groupName, machineID, machineIP string, webServerBind string) *GruntRegistry {
	myIP := net.ParseIP(machineIP)
	// myType := "leader"

	webServerRouter := mux.NewRouter().StrictSlash(true)

	// Create Run Context
	dockerClient, err := docker.NewClientFromEnv()
	utils.HandleError(err)
	runContext := types.RunContext{MyMachineID: machineID, MyIP: myIP, MyGroup: groupName, WebServerBind: webServerBind, WebServerRouter: webServerRouter, DockerClient: dockerClient}
	// Register and run
	gruntRegistry := NewGruntRegistry(&runContext)

	// Initialize workers
	store := CreateStore(&runContext, connectionString, groupName)
	api := &SporeAPI{}
	webserver := &WebServer{}
	planner := &Planner{}
	dockerRunner := &DockerRunner{}
	runContext.Store = store

	gruntRegistry.Start(store, api, webserver, planner, dockerRunner)
	return gruntRegistry
}
