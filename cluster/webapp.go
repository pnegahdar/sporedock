package cluster

import "github.com/samalba/dockerclient"

type WebApp struct {
	Count        int      `flatten:"{{ .ID }}/Count"`
	Env          string   `flatten:"{{ .ID }}/Env/"`
	ID           string   `flatten:"{{ .ID }}"`
	Image        string   `flatten:"{{ .ID }}/Image"`
	Tag          string   `flatten:"{{ .ID }}/Tag"`
	WebEndpoints []string `flatten:"{{ .ID }}/WebEndpoints/"` // Todo(parham) ensure uniqueness
	Weight       float32  `flatten:"{{ .ID }}/Weight"`
}

func (wa WebApp) ContainerConfig() dockerclient.ContainerConfig {
	return dockerclient.ContainerConfig{}
}
func (wa WebApp) GetImage() string {
	return wa.Image
}

func (wa WebApp) GetTag() string {
	return wa.Tag
}
func (wa WebApp) GetName() string {
	return wa.ID
}

type WebApps []WebApp

// Define the interface for sorting
func (wa WebApps) Len() int           { return len(wa) }
func (wa WebApps) Swap(i, j int)      { wa[i], wa[j] = wa[j], wa[i] }
func (wa WebApps) Less(i, j int) bool { return wa[i].Weight < wa[j].Weight }
