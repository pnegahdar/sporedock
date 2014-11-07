package cluster

import (
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/utils"
)

type WebApp struct {
	Count        int      `flatten:"{{ .ID }}/Count"`
	Env          string   `flatten:"{{ .ID }}/Env/"`
	ID           string   `flatten:"{{ .ID }}"`
	Image        string   `flatten:"{{ .ID }}/Image"`
	WebEndpoints []string `flatten:"{{ .ID }}/WebEndpoints/"`
	Weight       float32  `flatten:"{{ .ID }}/Weight"`
}

type WebApps []WebApp

// Define the interface for sorting
func (wa WebApps) Len() int           { return len(wa) }
func (wa WebApps) Swap(i, j int)      { wa[i], wa[j] = wa[j], wa[i] }
func (wa WebApps) Less(i, j int) bool { return wa[i].Weight < wa[j].Weight }

func (wa WebApps) Set() {
	data_json, err := marshall(wa)
	utils.HandleError(err)
	_, err1 := server.EtcdClient().Set(WebAppsKey, data_json, 0)
	utils.HandleError(err1)
}

func (wa *WebApps) Get() {
	etcd_resp, err := server.EtcdClient().Get(WebAppsKey, false, false)
	utils.HandleError(err)
	unmarshall(etcd_resp.Node.Value, wa)
}
