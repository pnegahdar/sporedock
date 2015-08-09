package cluster

import (
	"fmt"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"github.com/samalba/dockerclient"
)

type App struct {
	Count                   int
	Scheduler               string
	AttachedEnvs            []string
	ExtraEnv                map[string]string
	Tags                    map[string]string
	ID                      string
	Image                   string
	BalancedInternalTCPPort int
	Cpus                    int
	Memory                  int
}

func (wa App) RestartPolicy() dockerclient.RestartPolicy {
	policyName := fmt.Sprintf("SporedockRestartPolicy%vImage%v", wa.ID, wa.Image)
	restartPolicy := dockerclient.RestartPolicy{
		Name:              policyName,
		MaximumRetryCount: 5,
	}
	return restartPolicy
}

func (wa App) HostConfig() dockerclient.HostConfig {
	return dockerclient.HostConfig{
		PortBindings:  wa.PortBindings(),
		RestartPolicy: wa.RestartPolicy(),
	}
}

func (wa App) PortBindings() map[string][]dockerclient.PortBinding {
	anyPort := dockerclient.PortBinding{HostPort: "0"}
	bindings := map[string][]dockerclient.PortBinding{}
	bindings[fmt.Sprintf("%v/tcp", wa.BalancedInternalTCPPort)] = []dockerclient.PortBinding{anyPort}
	return bindings
}

func (wa App) Env() map[string]string {
	envList := []map[string]string{}
	for _, env := range wa.AttachedEnvs {
		envList = append(envList, FindEnv(env).Env)
	}
	envList = append(envList, wa.ExtraEnv)
	return utils.FlattenHashes(envList...)
}

func (wa App) ContainerConfig() dockerclient.ContainerConfig {
	envsForDocker := EnvAsDockerKV(wa.Env())
	exposedPorts := map[string]struct{}{}
	exposedPorts[fmt.Sprintf("%v/tcp", wa.BalancedInternalTCPPort)] = struct{}{}
	return dockerclient.ContainerConfig{
		Env:          envsForDocker,
		Image:        wa.Image,
		ExposedPorts: exposedPorts}
}

func (wa *App) Validate(rc *types.RunContext) error {
	return nil
}

func (wa *App) GetID() string {
	return wa.ID
}
