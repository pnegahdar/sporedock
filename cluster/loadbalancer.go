package cluster

import (
	"github.com/pnegahdar/sporedock/discovery"
	"github.com/pnegahdar/sporedock/utils"
)

const loadBalancerLocationsKey = "sporedock:loadbalancer:locations"
const loadBalancerLocationsLogLength = 200

type LoadBalancerLocations map[string]WebApp

func (lbl LoadBalancerLocations) Push() {
	data_json, err := marshall(lbl)
	utils.HandleError(err)
	store := discovery.GetStore()
	err = store.SetKeyWithLog(loadBalancerLocationsKey, data_json, loadBalancerLocationsLogLength)
	utils.HandleError(err)
}

func (lbl *LoadBalancerLocations) Pull() {
	store := discovery.GetStore()
	resp, err := store.GetKey(loadBalancerLocationsKey)
	utils.HandleError(err)
	unmarshall(resp, lbl)
}

func GetCurrentLBSet() Cluster {
	lb := LoadBalancerLocations{}
	lb.Pull()
	return lb
}
