package cluster

import (
	"github.com/pnegahdar/sporedock/discovery"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/utils"
)

type WorkerApp struct {
	Count  int     `flatten:"{{ .ID }}/Count"`
	Env    string  `flatten:"{{ .ID }}/Env/"`
	ID     string  `flatten:"{{ .ID }}"`
	Image  string  `flatten:"{{ .ID }}/Image"`
	Weight float32 `flatten:"{{ .ID }}/Weight"`
}

type WorkerApps []WorkerApp

// Define the interface for sorting
func (wa WorkerApps) Len() int           { return len(wa) }
func (wa WorkerApps) Swap(i, j int)      { wa[i], wa[j] = wa[j], wa[i] }
func (wa WorkerApps) Less(i, j int) bool { return wa[i].Weight < wa[j].Weight }

func (wa WorkerApps) EtcdSet() {
	data_json, err := marshall(wa)
	utils.HandleError(err)
	_, err1 := server.EtcdClient().Set(WorkerAppsKey, data_json, 0)
	utils.HandleError(err1)
}

func (wa *WorkerApps) EtcdGet() {
	etcd_resp, err := server.EtcdClient().Get(WorkerAppsKey, false, false)
	utils.HandleError(err)
	unmarshall(etcd_resp.Node.Value, wa)
}

type WorkerAppManifest struct {
	App     WorkerApp
	Machine discovery.Machine
}

type WorkerAppsManifests []WorkerAppManifest

func (w WorkerAppsManifests) Build() {

}
func (w WorkerAppsManifests) Set() {
	data_json, err := marshall(w)
	utils.HandleError(err)
	_, err1 := server.EtcdClient().Set(WorkerAppManifestsKey, data_json, 0)
	utils.HandleError(err1)
}

func (w *WorkerAppsManifests) Get() {
	etcd_resp, err := server.EtcdClient().Get(WorkerAppManifestsKey, false, false)
	utils.HandleError(err)
	unmarshall(etcd_resp.Node.Value, w)
}
