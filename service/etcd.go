package service

import (
	"encoding/json"
	etcdservice "github.com/coreos/etcd"
	etcdserviceconfig "github.com/coreos/etcd/config"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/pnegahdar/sporedock/settings"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
	"net/http"
)

var EtcdClientInstance etcdclient.Client
var EtcdServiceInstance etcdservice.Client

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

type EtcdService struct {
	Client  *etcdclient.Client
	Service *etcdservice.Client
}

func (s EtcdService) Init(Config) {
	if EtcdClientInstance == nil {
		EtcdClientInstance = etcdclient.NewClient(getDiscoveryPeers())
	}
	if EtcdServiceInstance == nil {
		config := etcdserviceconfig.New()
		config.Name = settings.GetInstanceName()
		config.DataDir = settings.GetEtcdDataDir()
		config.Discovery = settings.GetDiscoveryString()
		EtcdServiceInstance = etcdservice.New(config)
	}

	s.Client = *EtcdClientInstance
	s.Service = *EtcdServiceInstance
}

func (s EtcdService) Run() {
	s.Service.Run()
}
