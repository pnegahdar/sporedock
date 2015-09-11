package modules

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/pnegahdar/sporedock/cluster"
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
		d.client = client
		d.runContext = runContext
		utils.HandleError(err)
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
	plan, err := cluster.CurrentPlan(d.runContext)
	if err == types.ErrNoneFound {
		return
	}
	utils.HandleError(err)
	myJobs := plan.SporeSchedule[cluster.SporeID(d.runContext.MyMachineID)]
	guidsToKeep := []cluster.RunGuid{}
	for runGuid, app := range myJobs {
		cluster.PullApp(d.runContext, app)
		cluster.RunApp(d.runContext, runGuid, app)
		guidsToKeep = append(guidsToKeep, runGuid)
	}
	cluster.CleanupRemovedApps(d.runContext, guidsToKeep)
	cluster.CleanDeadApps(d.runContext)
}
