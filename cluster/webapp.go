package cluster

import (
	"fmt"
	"github.com/samalba/dockerclient"
)

type WebApp struct {
	Count        int
	AttachedEnvs []*Env
	ExtraEnv     map[string]string
	Tags         map[string]string
	ID           string
	Image        string
	WebEndpoints []string
}

func (wa WebApp) RestartPolicy() dockerclient.RestartPolicy {
	policyName := fmt.Sprintf("SporedockRestartPolicy%vImage%vTag%v", wa.ID, wa.Image, wa.Tag)
	restartPolicy := dockerclient.RestartPolicy{
		Name:              policyName,
		MaximumRetryCount: 5,
	}
	return restartPolicy
}

func (wa WebApp) HostConfig() dockerclient.HostConfig {
	return dockerclient.HostConfig{
		PortBindings:  wa.PortBindings(),
		RestartPolicy: wa.RestartPolicy(),
	}
}

func (wa WebApp) PortBindings() map[string][]dockerclient.PortBinding {
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
	exposedPorts["80/tcp"] = struct{}{}
	return dockerclient.ContainerConfig{
		Env:          envList,
		Image:        imageFull,
		ExposedPorts: exposedPorts}
}
func (wa WebApp) Image() string {
	return wa.Image
}

func (wa WebApp) Name() string {
	return fmt.Sprintf("Sporedock%v%v%v", wa.ID, wa.Image, wa.Tag)
}
