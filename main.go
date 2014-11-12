package main

import (
	logging "github.com/op/go-logging"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/director"
	"github.com/pnegahdar/sporedock/loadbalancer"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/settings"
	"time"
	"runtime"
)

func runStore() {
	go server.EtcdServer().Run()
	<-server.EtcdServer().ReadyNotify()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logging.SetLevel(logging.INFO, "main")
	settings.SetDiscoveryString("https://discovery.etcd.io/571c7c3a1d119ad5c75921c1d3d0a4a6")
	runStore()
	var c cluster.Cluster
	c.Import("sample_cluster.json")
	c.Push()
	go director.Direct()
	go loadbalancer.Serve()
	for {
		time.Sleep(1)
	}

}
