package cluster

import (
	"fmt"
	"github.com/pnegahdar/sporedock/discovery"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/utils"
	"sort"
)

type MachineManifest struct {
	Machine     discovery.Machine
	WebApps     WebApps
	WorkerApps  WorkerApps
	TotalWeight float32
}

type Manifests []MachineManifest

func (pm Manifests) Len() int           { return len(pm) }
func (pm Manifests) Swap(i, j int)      { pm[i], pm[j] = pm[j], pm[i] }
func (pm Manifests) Less(i, j int) bool { return pm[i].TotalWeight < pm[j].TotalWeight }

func (ms Manifests) Set() {
	data_json, err := marshall(ms)
	utils.HandleError(err)
	_, err1 := server.EtcdClient().CreateInOrder(ManifestLogsKey, data_json, 0)
	utils.HandleError(err1)
	_, err2 := server.EtcdClient().Set(CurrentManifestKey, data_json, 0)
	utils.HandleError(err2)
}

func (ms *Manifests) Get() {
	etcd_resp, err := server.EtcdClient().Get(CurrentManifestKey, false, false)
	utils.HandleError(err)
	unmarshall(etcd_resp.Node.Value, ms)
}

func (ms *Manifests) MyManifest(myMachine discovery.Machine) MachineManifest {
	for _, v := range(*ms) {
		if v.Machine == myMachine {
			return v
		}
	}
	panic("machine not found.")
}

func BuildClusterManifest(cluster Cluster) Manifests {
	utils.LogInfo(fmt.Sprintf("CLUSTER: %v", cluster))
	return buildAppManifests(cluster.WebApps, cluster.WorkerApps)
}

func buildAppManifests(webapps WebApps, workerapps WorkerApps) Manifests {
	webappsCopy, workerappsCopy := make(WebApps, len(webapps)), make(WorkerApps, len(webapps))

	copy(webappsCopy, webapps)
	copy(workerappsCopy, workerapps)
	machines := discovery.ListMachines()
	sort.Sort(webapps)
	sort.Sort(workerapps)
	var manifests Manifests
	utils.LogInfo(fmt.Sprintf("%v", machines))
	for _, x := range machines {
		packedMachine := MachineManifest{Machine: x}
		manifests = append(manifests, packedMachine)
	}
	utils.LogInfo(fmt.Sprintf("%v %v", len(workerapps), len(webapps)))
	utils.LogInfo(fmt.Sprintf("%v %v", len(webappsCopy), len(workerappsCopy)))
	for i := 0; i < len(webappsCopy)+len(workerappsCopy); i++ {
		utils.LogInfo("HERE")
		if webappsCopy[0].Weight > workerappsCopy[0].Weight {
			addWebApp(webappsCopy[0], manifests)
			webappsCopy = webappsCopy[1:]
		} else {
			addWorkerApp(workerappsCopy[0], manifests)
			workerappsCopy = workerappsCopy[1:]
		}
	}
	return manifests
}

func addWebApp(webapp WebApp, machines Manifests) {
	utils.LogInfo(fmt.Sprintf("Adding webapp %v", webapp.ID))
	count := webapp.Count
	if count > len(machines) {
		utils.LogWarn(fmt.Sprintf("App %v has a count %v that is greater than the machine count %v. Using machine count instead.", webapp.ID, webapp.Count, len(machines)))
		count = len(machines)
	}
	sort.Sort(machines)
	for i := 0; i < count; i++ {
		machines[i].WebApps = append(machines[i].WebApps, webapp)
		machines[i].TotalWeight = machines[i].TotalWeight + webapp.Weight
	}

}

func addWorkerApp(worker WorkerApp, machines Manifests) {
	utils.LogInfo(fmt.Sprintf("Adding workerapp %v", worker.ID))
	count := worker.Count
	if count > len(machines) {
		utils.LogWarn(fmt.Sprintf("App %v has a count %v that is greater than the machine count %v. Using machine count instead.", worker.ID, worker.Count, len(machines)))
		count = len(machines)
	}
	sort.Sort(machines)
	for i := 0; i < count; i++ {
		machines[i].WorkerApps = append(machines[i].WorkerApps, worker)
		machines[i].TotalWeight = machines[i].TotalWeight + worker.Weight
	}
}
