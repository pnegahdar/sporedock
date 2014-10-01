package etcd

import (
	"github.com/coreos/etcd/config"
	"github.com/coreos/etcd/etcd"
	"github.com/pnegahdar/sporedock/settings"
)

func Run() {
	config := config.New()
	config.Name = settings.GetInstanceName()
	config.DataDir = settings.GetEtcdDataDir()
	config.Discovery = settings.GetDiscoveryString()
	etcd := etcd.New(config)
	etcd.Run()
}
