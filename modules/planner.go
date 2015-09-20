package modules

import (
	"fmt"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sort"
	"sync"
	"time"
)

var PlanDebounceInterval = time.Second * 5

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
	if err == types.ErrNoneFound {
		return
	}
	sort.Sort(types.Apps(allApps))
	for _, app := range allApps {
		app.CountRemaining = app.Count
		scheduled := false
		// Todo: exclude repeat
		for _, schedulerFunc := range types.Schedulers {
			scheduled, err = schedulerFunc(&app, runContext, currentPlan, newPlan)
			if err != nil {
				types.HandleSchedulerError(err, app.ID, fmt.Sprintf("%v", schedulerFunc))
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
	appMeta, err := types.NewMeta(types.App{})
	utils.HandleError(err)
	anyAppEvent := types.StoreEvent(types.StorageActionAll, appMeta)
	eventList := []types.Event{types.EventStoreSporeAdded, types.EventStoreSporeExit, anyAppEvent}
	eventMessage := runContext.EventManager.ListenDebounced(runContext, &pl.stopCast, PlanDebounceInterval, eventList...)
	plan := func() {
		amLeader, err := types.AmLeader(runContext)
		if err == types.ErrNoneFound {
			return
		}
		utils.HandleError(err)
		if amLeader {
			pl.Plan(runContext)
		}
	}
	for {
		select {
		case <-eventMessage:
			utils.LogInfoF("Running %v", pl.ProcName())
			plan()
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
