package main

import (
	"fmt"
	logging "github.com/op/go-logging"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/discovery"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/settings"
	"time"
)

func setUp() {
	go server.EtcdServer().Run()
	<-server.EtcdServer().ReadyNotify()
}

func main() {
	settings.SetDiscoveryString("https://discovery.etcd.io/571c7c3a1d119ad5c75921c1d3d0a4a6")
	setUp()
	fmt.Println(discovery.CurrentMachine())
	logging.SetLevel(logging.DEBUG, "main")
	var c config.Cluster
	c.Import("sample_cluster.json")
	for {
		time.Sleep(1)
	}
}
