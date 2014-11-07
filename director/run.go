package director

import (
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/discovery"
	"github.com/pnegahdar/sporedock/settings"
	"time"
	"github.com/pnegahdar/sporedock/utils"
)

func Direct() {
	for {
		time.Sleep(settings.RebuildDelayS * time.Second)
		machine := discovery.CurrentMachine()
		if machine.State == "leader" {
			DistributeWork()
			SetupApps()
		} else {
			SetupApps()
		}
	}

}

func SetupApps() {
	utils.LogInfo("Setting up apps.")

}

func DistributeWork() {
	utils.LogInfo("Distributing work.")
	currentCluster := cluster.Cluster{}
	currentCluster.Get()
	pack := cluster.BuildClusterManifest(currentCluster)
	pack.Set()
}
