package sporedock

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
	settings.SetDiscoveryString("https://discovery.etcd.io/185a024ca8fc54d755680a2ffba5e183")
	server.RunAndWaitForEtcdServer()
	var c cluster.Cluster
	c.Import("sample_cluster.json")
	c.Push()
	director.Direct()
}
