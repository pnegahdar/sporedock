package etcd

import (
	"encoding/json"
	"github.com/coreos/go-etcd/etcd"
	"github.com/pnegahdar/sporedock/utils"
	"github.com/pnegahdar/sporedock/settings"
	"io/ioutil"
	"net/http"
)

var EtcdClient *etcd.Client

func getDiscoveryPeers() []string {
	resp, err := http.Get(settings.GetDiscoveryString())
	utils.HandleError(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	utils.HandleError(err)
	var data etcd.Response
	err = json.Unmarshal(body, &data)
	utils.HandleError(err)
	var peers []string
	for _, v := range data.Node.Nodes {
		peers = append(peers, v.Value)
	}
	return peers
}

func createClient() *etcd.Client {
	return etcd.NewClient(getDiscoveryPeers())
}

func GetClient() *etcd.Client {
	if EtcdClient == nil {
		EtcdClient = createClient()
	}
	return EtcdClient
}

func ListMachines() {
}

func MyMachine() {
}

func GetLeader() {
}

func AmLeader() {

}
