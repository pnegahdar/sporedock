package grunts

import (
	"fmt"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sort"
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
	allApps, err := cluster.AllApps(runContext)
	sort.Sort(cluster.Apps(allApps))
	if err == types.ErrNoneFound {
		return
	}
	utils.HandleError(err)
	allSpores, err := cluster.AllSporesMap(runContext)
	if err == types.ErrNoneFound {
		return
	}
	utils.HandleError(err)
	currentPlan, err := cluster.CurrentPlan(runContext)
	if err == types.ErrNoneFound {
		currentPlan = nil
	} else {
		utils.HandleError(err)
	}
	newPlan := &cluster.Plan{Spores: allSpores}
	for _, app := range allApps {
		scheduled := false
		for _, fn := range cluster.Schedulers {
			done, err := fn(&app, runContext, currentPlan, newPlan)
			if err != nil {
				cluster.HandleSchedulerError(err, app.ID, fmt.Sprintf("%v", fn))
			}
			if done {
				scheduled = done
				break
			}
		}
		if !scheduled {
			err := cluster.FinalScheduler(&app, runContext, currentPlan, newPlan)
			if err != nil {
				cluster.HandleSchedulerError(err, app.ID, "FinalScheduler")
			}
		}
	}
	cluster.SavePlan(runContext, newPlan)
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
