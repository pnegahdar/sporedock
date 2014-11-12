package cluster

import (
	"fmt"
	"github.com/samalba/dockerclient"
)

type WebApp struct {
	Count        int      `flatten:"{{ .ID }}/Count"`
	Env          string   `flatten:"{{ .ID }}/Env/"`
	ID           string   `flatten:"{{ .ID }}"`
	Image        string   `flatten:"{{ .ID }}/Image"`
	Tag          string   `flatten:"{{ .ID }}/Tag"`
	WebEndpoints []string `flatten:"{{ .ID }}/WebEndpoints/"` // Todo(parham) ensure uniqueness
	Weight       float32  `flatten:"{{ .ID }}/Weight"`
}

func (wa WebApp) GetRestartPolicy() dockerclient.RestartPolicy {
	policyName := fmt.Sprintf("SporedockRestartPolicy%vImage%vTag%v", wa.ID, wa.Image, wa.Tag)
	restartPolicy := dockerclient.RestartPolicy{
		Name:              policyName,
		MaximumRetryCount: 5,
	}
	return restartPolicy
}

func (wa WebApp) HostConfig() dockerclient.HostConfig {
	return dockerclient.HostConfig{
		PortBindings:  wa.GetPortBindings(),
		RestartPolicy: wa.GetRestartPolicy(),
	}
}

func (wa WebApp) GetPortBindings() map[string][]dockerclient.PortBinding {
	anyPort := dockerclient.PortBinding{HostPort: "0"}
	bindings := map[string][]dockerclient.PortBinding{}
	bindings["80/tcp"] = []dockerclient.PortBinding{anyPort}
	return bindings
}

func (wa WebApp) ContainerConfig() dockerclient.ContainerConfig {
	currentCluster := GetCurrentCluster()
	envList := currentCluster.GetEnv(wa.Env).AsDockerSlice()
	imageFull := fmt.Sprintf("%v:%v", wa.Image, wa.Tag)
	exposedPorts := map[string]struct{}{}
	exposedPorts["tcp/80"] = struct{}{}
	return dockerclient.ContainerConfig{
		Env:   envList,
		Image: imageFull,
		ExposedPorts: exposedPorts}
}
func (wa WebApp) GetImage() string {
	return wa.Image
}

func (wa WebApp) GetTag() string {
	return wa.Tag
}
func (wa WebApp) GetName() string {
	return fmt.Sprintf("Sporedock%v%v%v", wa.ID, wa.Image, wa.Tag)
}

type WebApps []WebApp

// Define the interface for sorting
func (wa WebApps) Len() int           { return len(wa) }
func (wa WebApps) Swap(i, j int)      { wa[i], wa[j] = wa[j], wa[i] }
func (wa WebApps) Less(i, j int) bool { return wa[i].Weight < wa[j].Weight }
