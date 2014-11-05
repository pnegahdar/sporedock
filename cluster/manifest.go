package cluster

import (
	"encoding/json"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/utils"
)

type Manifest struct {
	Machines  []MachineManifest
	timestamp int
}

type MachineManifest struct {
	MachineId  string
	WebApps    []WebApp
	WorkerApps []WorkerApp
}

func (m *Manifest) Distribute(cluster Cluster) {
	// TODO(parham)
}

func (m Manifest) Marshall() (string, error) {
	resp, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(resp[:]), nil
}

func (m *Manifest) UnMarshall(data string) error {
	err := json.Unmarshal([]byte(data), m)
	if err != nil {
		return err
	}
	return nil
}

func (m Manifest) EtcdSet() {
	data_json, err := m.Marshall()
	utils.HandleError(err)
	_, err1 := server.EtcdClient().CreateInOrder(ETCD_MANIFESTS_KEY, data_json, 0)
	utils.HandleError(err1)
	_, err2 := server.EtcdClient().Set(ETCD_CURRENT_MANIFEST_KEY, data_json, 0)
	utils.HandleError(err2)
}

func (c *Cluster) EtcdGet() {
	etcd_resp, err := server.EtcdClient().Get(ETCD_CURRENT_MANIFEST_KEY, false, false)
	utils.HandleError(err)
	c.UnMarshall(etcd_resp.Node.Value)
}
