package cluster

import (
	"github.com/pnegahdar/sporedock/discovery"
	"github.com/pnegahdar/sporedock/utils"
    "github.com/pnegahdar/sporedock/container"
)

const loadBalancerLocationsKey = "sporedock:loadbalancer:locations"
const loadBalancerLocationsLogLength = 200

type LoadBalancerLocations map[string]WebApp

func (lbl LoadBalancerLocations) Push() {
	data_json, err := marshall(lbl)
	utils.HandleError(err)
	store := store.GetStore()
	err = store.SetKeyWithLog(loadBalancerLocationsKey, data_json, loadBalancerLocationsLogLength)
	utils.HandleError(err)
}

func (lbl *LoadBalancerLocations) Pull() {
	store := store.GetStore()
	resp, err := store.GetKey(loadBalancerLocationsKey)
	utils.HandleError(err)
	unmarshall(resp, lbl)
}

func GetCurrentLBSet() Cluster {
	lb := LoadBalancerLocations{}
	lb.Pull()
	return lb
}

func AddRoute(app container.WebApp, )
