package server

import (
	"encoding/json"
	etcdclient "github.com/coreos/go-etcd/etcd"
	"github.com/coreos/etcd/etcdserver"
	"github.com/pnegahdar/sporedock/settings"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var etcdClient *etcdclient.Client
var etcdServer *etcdserver.EtcdServer

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

func EtcdServer() *etcdserver.EtcdServer {
	if etcdServer == nil {
		client, err := url.Parse("127.0.0.1:4001")
		peer, err := url.Parse("127.0.0.1:7001")
		utils.HandleError(err)
		cfg := &etcdserver.ServerConfig{
			Name:            settings.GetInstanceName(),
			ClientURLs:      []url.URL{*client},
			PeerURLs:        []url.URL{*peer},
			DataDir:         settings.GetEtcdDataDir(),
			DiscoveryURL:    settings.GetDiscoveryString(),
		}
		err = cfg.VerifyBootstrapConfig()
		utils.HandleError(err)
		var s *etcdserver.EtcdServer
		s, err = etcdserver.NewServer(cfg)
		utils.HandleError(err)
		etcdServer = s
	}
	return etcdServer
}

func RunAndWaitForEtcdServer() {
	go EtcdServer().Start()
	time.Sleep(time.Second * 5)
}
