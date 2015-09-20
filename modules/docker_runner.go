package modules

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sync"
)

type DockerRunner struct {
	initOnce   sync.Once
	client     *docker.Client
	stopCast   utils.SignalCast
	runContext *types.RunContext
}

func (d *DockerRunner) Init(runContext *types.RunContext) {
	d.initOnce.Do(func() {
		client, err := docker.NewClientFromEnv()
		utils.HandleError(err)
		runContext.Lock()
		runContext.DockerClient = client
		runContext.Unlock()
		d.runContext = runContext
	})
}

func (d DockerRunner) ProcName() string {
	return "DockerRunner"
}

func (d *DockerRunner) Stop() {
	d.stopCast.Signal()
}

func (d *DockerRunner) Run(runContext *types.RunContext) {
	exit, _ := d.stopCast.Listen()
	planMeta, err := types.NewMeta(&types.Plan{})
	utils.HandleError(err)
	listenForAppChange := types.StoreEvent(types.StorageActionAll, planMeta)
	eventMessage := runContext.EventManager.ListenDebounced(runContext, &d.stopCast, PlanDebounceInterval, listenForAppChange)
	for {
		select {
		case <-eventMessage:
			utils.LogInfoF("Running %v", d.ProcName())
			d.run()
		case <-exit:
			return
		}
	}
}

func (d *DockerRunner) ShouldRun(runContext *types.RunContext) bool {
	return true
}

func (d *DockerRunner) run() {
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
