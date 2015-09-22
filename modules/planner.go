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

type PlannerModule struct {
	sync.Mutex
	stopCast   utils.SignalCast
	stopCastMu sync.Mutex
	runContext *types.RunContext
}

func (plm PlannerModule) ShouldRun(runContext *types.RunContext) bool {
	//TODO: Master only
	return true
}

func (plm PlannerModule) ProcName() string {
	return "Planner"
}

func (plm *PlannerModule) Plan(runContext *types.RunContext) {
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

func (plm *PlannerModule) Run(runContext *types.RunContext) {
	exit, _ := plm.stopCast.Listen()
	appMeta, err := types.NewMeta(types.App{})
	utils.HandleError(err)
	anyAppEvent := types.StoreEvent(types.StorageActionAll, appMeta)
	eventList := []types.Event{types.EventStoreSporeAdded, types.EventStoreSporeExit, anyAppEvent}
	eventMessage := runContext.EventManager.ListenDebounced(runContext, &plm.stopCast, PlanDebounceInterval, eventList...)
	plm.run()
	for {
		select {
		case <-eventMessage:
			plm.run()
		case <-exit:
			return
		}
	}
}

func (plm *PlannerModule) run() {
	utils.LogInfoF("Running %v", plm.ProcName())
	amLeader, err := types.AmLeader(plm.runContext)
	if err == types.ErrNoneFound {
		return
	}
	utils.HandleError(err)
	if amLeader {
		plm.Plan(plm.runContext)
	}
}

func (plm *PlannerModule) Init(runContext *types.RunContext) {
	plm.Mutex.Lock()
	defer plm.Mutex.Unlock()
	plm.runContext = runContext
}

func (pl *PlannerModule) Stop() {
	pl.stopCastMu.Lock()
	defer pl.stopCastMu.Unlock()
	pl.stopCast.Signal()
}
