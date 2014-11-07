package cluster

import (
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/utils"
)

type Env struct {
	Env map[string]string `flatten:"{{ .ID }}/{{ .KEY }}"`
	ID  string            `flatten:"{{ .ID }}/"`
}

type Envs []Env

func (e Envs) Set() {
	data_json, err := marshall(e)
	utils.HandleError(err)
	_, err1 := server.EtcdClient().Set(EnvsKey, data_json, 0)
	utils.HandleError(err1)
}

func (e *Envs) Get() {
	etcd_resp, err := server.EtcdClient().Get(EnvsKey, false, false)
	utils.HandleError(err)
	unmarshall(etcd_resp.Node.Value, e)
}
