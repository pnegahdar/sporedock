package main

import (
	logging "github.com/op/go-logging"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/director"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/settings"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logging.SetLevel(logging.INFO, "main")
	settings.SetDiscoveryString("https://discovery.etcd.io/4849358d5c86b81d18b698eb8dc921ea")
	server.RunAndWaitForEtcdServer()
	var c cluster.Cluster
	c.Import("sample_cluster.json")
	c.Push()
	director.Direct()
}
