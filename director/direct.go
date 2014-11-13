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
	CleanDeadApps()
	for {
		time.Sleep(settings.RebuildDelayS * time.Second)
		machine := discovery.CurrentMachine()
		PrepMyApps()
		if machine.State == "leader" {
			lastCluster = DistributeWork(lastCluster)
			CleanupLocations()

		}
		ProcessMyManifest()
	}
}

func PrepMyApps() {
	utils.LogInfo("Syncing cluster.")
	currentCluster := cluster.GetCurrentCluster()
	var waitGroup sync.WaitGroup
	for _, app := range currentCluster.IterApps() {
		waitGroup.Add(1)
		go pullApp(app, &waitGroup)
	}
	waitGroup.Wait()

}

func DistributeWork(lastCluster cluster.Cluster) cluster.Cluster {
	utils.LogInfo("Distributing work.")
	currentCluster := cluster.GetCurrentCluster()
	if !reflect.DeepEqual(currentCluster, lastCluster) {
		utils.LogInfo("Cluster change detected. Rebuilding.")
		pack := cluster.BuildClusterManifest(currentCluster)
		pack.Push()
	} else {
		utils.LogInfo("Cluster has not changed.")
	}
	return currentCluster
}

func ProcessMyManifest() {
	currentManifest := cluster.GetCurrentManifest()
	myManifest := currentManifest.MyManifest(discovery.CurrentMachine())
	apps := myManifest.IterApps()

	waitGroup := sync.WaitGroup{}
	appNames := []string{}
	for _, app := range apps {
		appNames = append(appNames, app.GetName())
		go RunAppSafe(app, myManifest, &waitGroup)
		waitGroup.Add(1)
	}
	waitGroup.Wait()
	CleanupRemovedApps(appNames)
	UpdateLocations(appNames)
}
