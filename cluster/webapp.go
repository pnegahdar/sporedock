package cluster

import (
	"encoding/json"
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

func (wa WebApps) EtcdSet() {
	data_json, err := marshall(wa)
	utils.HandleError(err)
	_, err1 := server.EtcdClient().Set(WebAppsKey, data_json, 0)
	utils.HandleError(err1)
}

func (wa *WebApps) EtcdGet() {
	etcd_resp, err := server.EtcdClient().Get(WebAppsKey, false, false)
	utils.HandleError(err)
	unmarshall(etcd_resp.Node.Value, wa)
}
