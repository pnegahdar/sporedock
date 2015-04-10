package grunt

import (
    "github.com/pnegahdar/sporedock/store"
    "github.com/pnegahdar/sporedock/utils"
    "fmt"
    "runtime/debug"
    "time"
    "errors"
)

const RestartDelaySeconds = 1

var ErrorAlreadyRegisteredGrunts = errors.New("Grunts have already been registered once.")

type RunContext struct {
     store store.SporeStore
}

type Grunt interface{
    Name() string
    Run(runContext RunContext)
    ShouldRun(mySpore store.Spore) bool

}

type GruntRegistry struct {
    Grunts map[string]Grunt
    runCount map[string]int
    context RunContext
    startMe chan string
}

func (gr GruntRegistry) RegisterGrunts(grunts ...Grunt) {
    if gr.startMe != nil{
        utils.HandleError(ErrorAlreadyRegisteredGrunts)
    }
    gr.startMe = make(chan string, len(grunts))
    for _, grunt := range (grunts) {
        gruntName := grunt.Name()
        utils.LogInfo(fmt.Sprintf("Adding grunt %v", gruntName))
        gr.Grunts[gruntName] = grunt
        gr.runCount[gruntName] = 0
        gr.startMe <- gruntName
    }

}

func (gr GruntRegistry) runGrunt(gruntName string) {
    grunt, exists := gr.Grunts[gruntName]
    if !exists {
        return
    }
    runCount, exists := gr.runCount[gruntName]
    if !exists {
        runCount = 0
    }
    delayTot := RestartDelaySeconds * runCount
    utils.LogInfo(fmt.Sprintf("Running grunt %v with delay of %v seconds", gruntName, delayTot))
    go func() {
        defer func() {
            if rec := recover(); rec != nil {
                utils.LogInfo(fmt.Sprintf("Grunt %v paniced", gruntName))
                debug.PrintStack()
            }
        }()
        time.Sleep(delayTot*time.Second)
        utils.LogInfo(fmt.Sprintf("Running grunt %v", gruntName))
        grunt.Run(gr.context)

        //Send over again
        utils.LogInfo(fmt.Sprintf("Grunt %v exited", gruntName))
        gr.startMe <- gruntName

    }()
}

func (gr GruntRegistry) Start() {
    for gruntToStart := range (gr.startMe) {
        go gr.runGrunt(gruntToStart)
    }
}

