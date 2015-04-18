package grunts

import (
	"github.com/pnegahdar/sporedock/utils"
	"time"
)

type TestRunner struct{}

func (tr TestRunner) ProcName() string {
	return "TEST RUNNER"
}

func (tr TestRunner) ShouldRun(runContext RunContext) bool {
	return true
}

func (tr TestRunner) Run(runContext *RunContext) {
	for {
		utils.LogInfo("SLEEP TEST!")
		time.Sleep(time.Second)
	}
}
