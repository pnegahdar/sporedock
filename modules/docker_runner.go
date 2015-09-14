package modules

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sync"
	"time"
)

var syncDockerEveryD = time.Millisecond * 1000

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
	for {
		select {
		case <-time.After(syncDockerEveryD):
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
		types.PullApp(d.runContext, app)
		types.RunApp(d.runContext, runGuid, app)
		guidsToKeep = append(guidsToKeep, runGuid)
	}
	types.CleanupRemovedApps(d.runContext, guidsToKeep)
	types.CleanDeadApps(d.runContext)
}
