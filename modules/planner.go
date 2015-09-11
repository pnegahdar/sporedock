package modules

import (
	"fmt"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sort"
	"sync"
	"time"
)

const PlanEveryMs = 5000

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
	currentPlan, err := types.CurrentPlan(runContext)
	if err == types.ErrNoneFound {
		currentPlan = nil
	} else {
		utils.HandleError(err)
	}
	newPlan, err := types.NewPlan(runContext)
	if err == types.ErrNoneFound {
		return
	}
	utils.HandleError(err)
	allApps, err := types.AllApps(runContext)
	sort.Sort(types.Apps(allApps))
	if err == types.ErrNoneFound {
		return
	}
	sort.Sort(types.Apps(allApps))
	for _, app := range allApps {
		app.CountRemaining = app.Count
		scheduled := false
		// Todo: exclude repeat
		for _, scheduler_fun := range types.Schedulers {
			scheduled, err = scheduler_fun(&app, runContext, currentPlan, newPlan)
			if err != nil {
				types.HandleSchedulerError(err, app.ID, fmt.Sprintf("%v", scheduler_fun))
			}
			if scheduled {
				break
			}
		}
		if !scheduled {
			err := types.FinalScheduler(&app, runContext, currentPlan, newPlan)
			if err != nil {
				types.HandleSchedulerError(err, app.ID, "FinalScheduler")
			}
		}
	}
	err = types.SavePlan(runContext, newPlan)
	utils.HandleError(err)
}

func (pl *Planner) Run(runContext *types.RunContext) {
	exit, _ := pl.stopCast.Listen()
	for {
		select {
		case <-time.After(time.Millisecond * PlanEveryMs):
			amLeader, err := types.AmLeader(runContext)
			if err == types.ErrNoneFound {
				continue
			}
			utils.HandleError(err)
			if amLeader {
				pl.Plan(runContext)
			}
		// Todo: Also bind the app create/delete event
		case <-exit:
			return
		}
	}
}

func (pl *Planner) Init(runContext *types.RunContext) {
	return
}

func (pl *Planner) Stop() {
	pl.stopCastMu.Lock()
	defer pl.stopCastMu.Unlock()
	pl.stopCast.Signal()
}
