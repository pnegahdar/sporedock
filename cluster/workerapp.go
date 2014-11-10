package cluster

import "github.com/samalba/dockerclient"

type WorkerApp struct {
	Count  int     `flatten:"{{ .ID }}/Count"`
	Env    string  `flatten:"{{ .ID }}/Env/"`
	ID     string  `flatten:"{{ .ID }}"`
	Image  string  `flatten:"{{ .ID }}/Image"`
	Tag    string  `flatten:"{{ .ID }}/Tag"`
	Weight float32 `flatten:"{{ .ID }}/Weight"`
}

func (wa WorkerApp) ContainerConfig() dockerclient.ContainerConfig {
	return dockerclient.ContainerConfig{}
}
func (wa WorkerApp) GetImage() string {
	return wa.Image
}

func (wa WorkerApp) GetTag() string {
	return wa.Tag
}
func (wa WorkerApp) GetName() string {
	return wa.ID
}

type WorkerApps []WorkerApp

// Define the interface for sorting
func (wa WorkerApps) Len() int           { return len(wa) }
func (wa WorkerApps) Swap(i, j int)      { wa[i], wa[j] = wa[j], wa[i] }
func (wa WorkerApps) Less(i, j int) bool { return wa[i].Weight < wa[j].Weight }
