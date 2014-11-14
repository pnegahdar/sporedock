package server

import (
	"encoding/json"
	etcdserviceconfig "github.com/coreos/etcd/config"
	etcdservice "github.com/coreos/etcd/etcd"
	etcdpeersclient "github.com/coreos/etcd/server"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/pnegahdar/sporedock/settings"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var etcdClient *etcdclient.Client
var etcdServer *etcdservice.Etcd
var etcdPeerClient *etcdpeersclient.Client

func getDiscoveryPeers() []string {
	resp, err := http.Get(settings.GetDiscoveryString())
	utils.HandleError(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	utils.HandleError(err)
	var data etcdclient.Response
	err = json.Unmarshal(body, &data)
	utils.HandleError(err)
	var peers []string
	for _, v := range data.Node.Nodes {
		parsedUrl, err := url.Parse(v.Value)
		utils.HandleError(err)
		peers = append(peers, "http://"+strings.Split(parsedUrl.Host, ":")[0]+":4001")
	}
	return peers
}

func EtcdClient() *etcdclient.Client {
	if etcdClient == nil {
		etcdClient = etcdclient.NewClient(getDiscoveryPeers())
	}
	return etcdClient

}
func EtcdServer() *etcdservice.Etcd {
	if etcdServer == nil {
		config := etcdserviceconfig.New()
		config.Name = settings.GetInstanceName()
		config.DataDir = settings.GetEtcdDataDir()
		config.Discovery = settings.GetDiscoveryString()
		etcdServer = etcdservice.New(config)
	}
	return etcdServer
}

func EtcdPeerClient() *etcdpeersclient.Client {
	if etcdPeerClient == nil {
		etcdPeerClient = etcdpeersclient.NewClient(&http.Transport{})
	}
	return etcdPeerClient
}

func RunAndWaitForEtcdServer() {
	go EtcdServer().Run()
	<-EtcdServer().ReadyNotify()
}
