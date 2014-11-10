package director

import (
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/discovery"
	"github.com/pnegahdar/sporedock/settings"
	"github.com/pnegahdar/sporedock/utils"
	"reflect"
	"sync"
	"time"
)

func Direct() {
	lastCluster := cluster.Cluster{}
	for {
		time.Sleep(settings.RebuildDelayS * time.Second)
		machine := discovery.CurrentMachine()
		PrepMyApps()
		if machine.State == "leader" {
			lastCluster = DistributeWork(lastCluster)
		}
	}
}

func PrepMyApps() {
	utils.LogInfo("Syncing cluster.")
	currentCluster := cluster.Cluster{}
	currentCluster.Pull()
	var waitGroup sync.WaitGroup
	for _, app := range currentCluster.IterApps() {
		go pullApp(app, &waitGroup)
	}
	waitGroup.Wait()

}

func DistributeWork(lastCluster cluster.Cluster) cluster.Cluster {
	utils.LogInfo("Distributing work.")
	currentCluster := cluster.Cluster{}
	currentCluster.Pull()

	if !reflect.DeepEqual(currentCluster, lastCluster) {
		utils.LogInfo("Cluster change detected. Rebuilding.")
		pack := cluster.BuildClusterManifest(currentCluster)
		pack.Push()
	} else {
		utils.LogInfo("Cluster has not changed.")
	}
	return currentCluster
}
