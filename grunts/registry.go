package grunts

import (
	"fmt"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"net"
	"time"
)

const RestartDecaySeconds = 1

type GruntRegistry struct {
	Grunts   map[string]types.Grunt
	Context  *types.RunContext
	runCount map[string]int
	startMe  chan string
	stopCast utils.SignalCast
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
	grunt, exists := gr.Grunts[gruntName]
	if !exists {
		utils.LogWarn(fmt.Sprintf("Grunt %v DNE %v", gruntName, grunt))
		return
	}
	runCount := gr.runCount[gruntName]
	delayTot := RestartDecaySeconds * runCount
	gr.runCount[gruntName] = runCount + 1
	utils.LogInfo(fmt.Sprintf("Running grunt %v with delay of %v seconds", gruntName, delayTot))
	stopChan := gr.stopCast.Listen()
	select {
	case <-time.After(time.Duration(delayTot) * time.Second):
		defer func() {
			if rec := recover(); rec != nil {
				utils.LogInfo(fmt.Sprintf("Grunt %v paniced", gruntName))
			} else {
				utils.LogInfo(fmt.Sprintf("Grunt %v exited", gruntName))
			}
			gr.startMe <- gruntName
		}()
		utils.LogInfo(fmt.Sprintf("Running grunt %v", gruntName))
		grunt.Run(gr.Context)
	case <-stopChan:
		return
	}
}

func (gr *GruntRegistry) Start(grunts ...types.Grunt) {
	gr.registerGrunts(grunts...)
	utils.LogInfo("Runner started.")
	// Range blocks on startMe channel
	go func() {
		stopChan := gr.stopCast.Listen()
		for {
			select {
			case gruntToStart := <-gr.startMe:
				go gr.runGrunt(gruntToStart)
			case <-stopChan:
				return
			}
		}
	}()
}

func (gr *GruntRegistry) Stop() {
	for _, grunt := range gr.Grunts {
		grunt.Stop()
	}
	gr.stopCast.Signal()

}

func (gr *GruntRegistry) Wait() {
	<-gr.stopCast.Listen()
}

func NewGruntRegistry(rc *types.RunContext) *GruntRegistry {
	grunts := make(map[string]types.Grunt)
	runCount := make(map[string]int)
	return &GruntRegistry{Context: rc, Grunts: grunts, runCount: runCount}
}

func CreateAndRun(connectionString, groupName, machineID, machineIP string) *GruntRegistry {
	myIP := net.ParseIP("127.0.0.1")
	// myType := "leader"

	// Create Run Context
	runContext := types.RunContext{MyMachineID: machineID, MyIP: myIP, MyGroup: groupName}
	// Register and run
	gruntRegistry := NewGruntRegistry(&runContext)

	// Initialize workers
	store := CreateStore(&runContext, connectionString, groupName)
	api := SporeAPI{}
	runContext.Store = store

	gruntRegistry.Start(store, api)
	return gruntRegistry
}
