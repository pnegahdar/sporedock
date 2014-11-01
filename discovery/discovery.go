package discovery

type Machine struct {
	UID  string
	ipv4 string
}

type SporeDockDiscovery interface {
	ListMachines() []Machine
	CurrentMachine() Machine
	GetLeader() Machine
	AmLeader() bool
}
