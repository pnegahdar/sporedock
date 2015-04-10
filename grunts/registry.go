package grunts

import (
//"github.com/pnegahdar/sporedock/store"
    "github.com/pnegahdar/sporedock/utils"
    "fmt"
    "time"
)

const RestartDelaySeconds = 1

type RunContext struct {
    //store store.SporeStore
}

type Grunt interface{
    Name() string
    Run(runContext RunContext)
    ShouldRun(runContext RunContext) bool

}

type GruntRegistry struct {
    Grunts map[string]Grunt
    Context RunContext
    runCount map[string]int
    startMe chan string
}

func (gr *GruntRegistry) registerGrunts(grunts ...Grunt) {
    gr.Grunts = make(map[string]Grunt)
    gr.runCount = make(map[string]int)
    gr.startMe = make(chan string, len(grunts))
    // Todo: check should run
    utils.LogInfo(fmt.Sprintf("%v grunts", len(grunts)))
    for _, grunt := range (grunts) {
        gruntName := grunt.Name()
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
    delayTot := RestartDelaySeconds * runCount
    gr.runCount[gruntName] = runCount + 1
    utils.LogInfo(fmt.Sprintf("Running grunt %v with delay of %v seconds", gruntName, delayTot))
    go func() {
        defer func() {
            if rec := recover(); rec != nil {
                utils.LogInfo(fmt.Sprintf("Grunt %v paniced", gruntName))
                gr.startMe <- gruntName
            }
        }()
        time.Sleep(time.Duration(delayTot)*time.Second)
        utils.LogInfo(fmt.Sprintf("Running grunt %v", gruntName))
        grunt.Run(gr.Context)

        //Send over again
        utils.LogInfo(fmt.Sprintf("Grunt %v exited", gruntName))
        gr.startMe <- gruntName

    }()
}

func (gr *GruntRegistry) run() {
    utils.LogInfo("Runner started.")
    for gruntToStart := range (gr.startMe) {
        go gr.runGrunt(gruntToStart)
    }
}

func (gr *GruntRegistry) Start(grunts ...Grunt) {
    gr.registerGrunts(grunts...)
    gr.run()
}

