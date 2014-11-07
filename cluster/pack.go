package cluster

import (
	"fmt"
	"github.com/pnegahdar/sporedock/discovery"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/utils"
	"sort"
)

type PackedMachine struct {
	Machine     discovery.Machine
	WebApps     WebApps
	WorkerApps  WorkerApps
	TotalWeight float32
}

type PackedMachines []PackedMachine

func (pm PackedMachines) Len() int           { return len(pm) }
func (pm PackedMachines) Swap(i, j int)      { pm[i], pm[j] = pm[j], pm[i] }
func (pm PackedMachines) Less(i, j int) bool { return pm[i].TotalWeight < pm[j].TotalWeight }

func PackApps(webapps WebApps, workerapps WorkerApps) PackedMachines {
	webappsCopy, workerappsCopy := make(WebApps, len(webapps)), make(WorkerApps, len(webapps))
	copy(webappsCopy, webapps)
	copy(workerappsCopy, workerapps)
	machines := discovery.ListMachines()
	sort.Sort(webapps)
	sort.Sort(workerapps)
	var packedMachines PackedMachines
	for _, x := range machines {
		packedMachine := PackedMachine{Machine: x}
		packedMachines = append(packedMachines, packedMachine)
	}
	for i := 0; i < len(webapps)+len(workerapps); i++ {
		if webappsCopy[0].Weight > workerappsCopy[0].Weight {
			addWebApp(webappsCopy[0], packedMachines)
			webappsCopy = webappsCopy[1:]
		} else {
			addWorkerApp(workerappsCopy[0], packedMachines)
			workerappsCopy = workerappsCopy[1:]
		}
	}
	return packedMachines
}

func addWebApp(webapp WebApp, machines PackedMachines) {
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

func addWorkerApp(worker WorkerApp, machines PackedMachines) {
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

func (pms PackedMachines) Set() {
	data_json, err := marshall(pms)
	utils.HandleError(err)
	_, err1 := server.EtcdClient().Set(WebAppsKey, data_json, 0)
	utils.HandleError(err1)
}

func (pms *PackedMachines) Get() {
	etcd_resp, err := server.EtcdClient().Get(WebAppsKey, false, false)
	utils.HandleError(err)
	unmarshall(etcd_resp.Node.Value, pms)
}
