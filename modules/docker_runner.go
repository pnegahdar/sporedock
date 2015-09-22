package modules

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sync"
)

type DockerRunnerModule struct {
	initOnce   sync.Once
	client     *docker.Client
	stopCast   utils.SignalCast
	runContext *types.RunContext
}

func (drm *DockerRunnerModule) Init(runContext *types.RunContext) {
	drm.initOnce.Do(func() {
		client, err := docker.NewClientFromEnv()
		utils.HandleError(err)
		runContext.Lock()
		runContext.DockerClient = client
		runContext.Unlock()
		drm.runContext = runContext
	})
}

func (drm DockerRunnerModule) ProcName() string {
	return "DockerRunner"
}

func (drm *DockerRunnerModule) Stop() {
	drm.stopCast.Signal()
}

func (drm *DockerRunnerModule) Run(runContext *types.RunContext) {
	exit, _ := drm.stopCast.Listen()
	planMeta, err := types.NewMeta(&types.Plan{})
	utils.HandleError(err)
	listenForAppChange := types.StoreEvent(types.StorageActionAll, planMeta)
	eventMessage := runContext.EventManager.ListenDebounced(runContext, &drm.stopCast, PlanDebounceInterval, listenForAppChange)
	drm.run()
	for {
		select {
		case <-eventMessage:
			drm.run()
		case <-exit:
			return
		}
	}
}

func (drm *DockerRunnerModule) ShouldRun(runContext *types.RunContext) bool {
	return true
}

func (d *DockerRunnerModule) run() {
	utils.LogInfoF("Running %v", d.ProcName())
	plan, err := types.CurrentPlan(d.runContext)
	if err == types.ErrNoneFound {
		return
	}
	utils.HandleError(err)
	myJobs := plan.SporeSchedule[types.SporeID(d.runContext.MyMachineID)]
	guidsToKeep := []types.RunGuid{}
	for runGuid, app := range myJobs {
		types.PullApp(d.runContext, &app)
		types.RunApp(d.runContext, runGuid, &app)
		guidsToKeep = append(guidsToKeep, runGuid)
	}
	types.CleanupRemovedApps(d.runContext, guidsToKeep)
	// Todo(parham): Delay n hours
	types.CleanDeadApps(d.runContext)
}
