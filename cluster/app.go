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
	PinSpore                string
	AttachedEnvs            []string
	ExtraEnv                map[string]string
	Tags                    map[string]string
	ID                      string
	Image                   string
	BalancedInternalTCPPort int
	types.Sizable
}

func (a App) Size() float64 {
	return types.GetSize(a.Cpus, a.Mem)
}

type Apps []App

func (a Apps) Len() int      { return len(a) }
func (a Apps) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Apps) Less(i, j int) bool {
	return types.GetSize(a[i].Cpus, a[i].Mem) < types.GetSize(a[j].Cpus, a[j].Mem)
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
	// Todo: cpus and memory
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

func AllApps(runContext *types.RunContext) ([]App, error) {
	apps := []App{}
	err := runContext.Store.GetAll(&apps, 0, types.SentinelEnd)
	if err != nil {
		return nil, err
	}
	return apps, nil
}
