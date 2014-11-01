package discovery

import (
	"github.com/pnegahdar/sporedock/service"
)

type EtcdDiscovery struct {
	service service.EtcdService
}
func (d EtcdDiscovery) ListMachines() []Machine {
	return []Machine{}
}

func (d EtcdDiscovery) CurrentMachine() Machine {
	return Machine{}
}
func (d EtcdDiscovery) GetLeader() Machine {
	return Machine{}
}
func (d EtcdDiscovery) AmLeader() bool {
	return true
}