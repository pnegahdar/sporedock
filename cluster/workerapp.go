package cluster

import "encoding/json"

type WorkerApp struct {
	Env   string `flatten:"{{ .ID }}/Env/"`
	ID    string `flatten:"{{ .ID }}"`
	Image string `flatten:"{{ .ID }}/Image"`
}

type WorkerApps []WorkerApp

func (wa) Marshall() (string, error) {
	resp, err := json.Marshal(wa)
	if err != nil {
		return "", err
	}
	return string(resp[:]), nil
}

func (e *Env) UnMarshall(data string) error {
	err := json.Unmarshal([]byte(data), e)
	if err != nil {
		return err
	}
	return nil
}

func (wa WorkerApp) getEtcdKey( string) string {
	return ETCD_ENV_PREFIX + wa.ID
}

func (wa WorkerApp) EtcdSet() {
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
