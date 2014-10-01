package app

import "github.com/pnegahdar/sporedock/etcd"

func StartServer() {
	etcd.Run()
}
