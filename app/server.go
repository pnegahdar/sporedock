package app

import (
	"github.com/pnegahdar/sporedock/etcd"
)

func StartServer(discovery_url string) {
	etcd.Run(discovery_url)
}
