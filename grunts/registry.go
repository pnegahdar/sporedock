package grunts

import (
    "fmt"
    "github.com/pnegahdar/sporedock/utils"
    "net"
    "time"
    "github.com/pnegahdar/sporedock/types"
)

const RestartDecaySeconds = 1

type Grunt interface {
    ProcName() string
    Run(runContext *types.RunContext)
    ShouldRun(runContext types.RunContext) bool
}

type GruntRegistry struct {
    Grunts   map[string]Grunt
    Context  *types.RunContext
    runCount map[string]int
    startMe  chan string
}

func (gr *GruntRegistry) registerGrunts(grunts ...Grunt) {
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
    go func() {
        defer func() {
            if rec := recover(); rec != nil {
                utils.LogInfo(fmt.Sprintf("Grunt %v paniced", gruntName))
                gr.startMe <- gruntName
            }
        }()
        time.Sleep(time.Duration(delayTot) * time.Second)
        utils.LogInfo(fmt.Sprintf("Running grunt %v", gruntName))
        grunt.Run(gr.Context)

        //Send over again
        utils.LogInfo(fmt.Sprintf("Grunt %v exited", gruntName))
        gr.startMe <- gruntName

    }()
}

func (gr *GruntRegistry) Start(grunts ...Grunt) {
    gr.registerGrunts(grunts...)
    utils.LogInfo("Runner started.")
    // Range blocks on startMe channel
    for gruntToStart := range gr.startMe {
        go gr.runGrunt(gruntToStart)
    }
}

func CreateAndRun() {
    connectionString := "redis://localhost:6379"
    groupName := "testGroup"
    machineID := "myMachine"
    myIP := net.ParseIP("127.0.0.1")
    // myType := "leader"

    // Initialize workers
    store := CreateStore(connectionString, groupName)
    api := SporeAPI{}

    // Create Run Context
    runContext := types.RunContext{MyMachineID: machineID, Store: store, MyIP: myIP, MyGroup: groupName}
    // Register and run
    grunts := make(map[string]Grunt)
    runCount := make(map[string]int)
    gruntRegistry := GruntRegistry{Context: &runContext, Grunts: grunts, runCount: runCount}
    gruntRegistry.Start(store, api)
}
