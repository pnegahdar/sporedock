package discovery

type Machine struct {
	UID  string
	ipv4 string
}

type SporeDockDiscovery interface {
	GetService()
	ListMachines() []Machine
	CurrentMachine() Machine
	GetLeader() Machine
	AmLeader() bool
	Run()
}
