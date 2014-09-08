package etcd

import(
	"github.com/coreos/etcd/config"
	"github.com/coreos/etcd/etcd"
)


func Run(discovery_url string){
	config := config.New()
	config.Discovery = discovery_url
	etcd := etcd.New(config)
	etcd.Run()
}
