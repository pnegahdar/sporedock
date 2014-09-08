package worker

import (
"github.com/pnegahdar/SporeDock/etcd"
)

func Run(discovery_url string){
	etcd.Run(discovery_url)
}
