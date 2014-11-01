package main

import (
	logging "github.com/op/go-logging"
	"github.com/pnegahdar/sporedock/config"
	"github.com/pnegahdar/sporedock/discovery"
	"github.com/pnegahdar/sporedock/store"
)

var Store store.SporeDockStore
var Discovery discovery.SporeDockDiscovery

func setModules() {
	Store = store.EtcdStore{}
	Discovery = discovery.EtcdDiscovery{}
}

func startModules(){
	Discovery.Run()
}

func main() {
	setModules()
	startModules()
	logging.SetLevel(logging.DEBUG, "main")
	config.ImportClusterConfigFromFile("sample_cluster.json")
}
