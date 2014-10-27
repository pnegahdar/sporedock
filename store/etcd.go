package store

import (
	"github.com/pnegahdar/sporedock/service"
)

type EtcdStore struct {
	service service.EtcdService
}

func (store EtcdStore) Get(string) string {
	return ""
}

func (store EtcdStore) Set(string) string {
	return ""
}

func (store EtcdStore) Exists(string) bool {
	return true
}
