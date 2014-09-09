package worker

import (
	"github.com/pnegahdar/sporedock/etcd"
)

func Run(discovery_url string) {
	etcd.Run(discovery_url)
}
