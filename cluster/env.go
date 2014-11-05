package cluster

import (
	"encoding/json"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/utils"
)

type Env struct {
	Env map[string]string `flatten:"{{ .ID }}/{{ .KEY }}"`
	ID  string            `flatten:"{{ .ID }}/"`
}

type Envs []Env

func (e Envs) Marshall() (string, error) {
	resp, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(resp[:]), nil
}

func (e *Envs) UnMarshall(data string) error {
	err := json.Unmarshal([]byte(data), e)
	if err != nil {
		return err
	}
	return nil
}

func (e Env) EtcdSet() {
	data_json, err1 := e.Marshall()
	utils.HandleError(err1)
	_, err1 := server.EtcdClient().Set(e.getEtcdKey(), data_json, 0)
	utils.HandleError(err1)
}

func (e *Env) EtcdGet() {
	etcd_resp, err := server.EtcdClient().Get(e.getEtcdKey(), false, false)
	utils.HandleError(err)
	e.UnMarshall(etcd_resp.Node.Value)
}
