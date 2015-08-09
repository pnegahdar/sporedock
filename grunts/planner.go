package grunts

import (
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sync"
	"time"
)

const PlanEveryMs = 3000

type Planner struct {
	sync.Mutex
	stopCast   utils.SignalCast
	stopCastMu sync.Mutex
}

func (pl Planner) ShouldRun(runContext *types.RunContext) bool {
	//TODO: Master only
	return true
}

func (pl Planner) ProcName() string {
	return "Planner"
}

func (pl *Planner) Plan(runContext *types.RunContext) {
}

func (pl *Planner) Run(runContext *types.RunContext) {
	pl.Lock()
	defer pl.Unlock()
	exit, _ := pl.stopCast.Listen()
	for {
		select {
		case <-time.After(time.Millisecond * PlanEveryMs):
			pl.Plan(runContext)
		case <-exit:
			return
		}
	}
}

func (pl *Planner) Stop() {
	pl.stopCastMu.Lock()
	defer pl.stopCastMu.Unlock()
	pl.stopCast.Signal()
}

type Assignment struct {
}

type PlanContext struct {
}
