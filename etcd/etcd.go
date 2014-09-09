package etcd

import (
	"github.com/coreos/etcd/config"
	"github.com/coreos/etcd/etcd"
	"github.com/pnegahdar/sporedock/settings"
)

func Run(discovery_url string) {
	config := config.New()
	config.Name = settings.GetInstanceName()
	config.Discovery = discovery_url
	etcd := etcd.New(config)
	etcd.Run()
}
