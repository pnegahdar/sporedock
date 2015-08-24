package cluster

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
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

func (wa App) DockerContainerOptions(runContext *types.RunContext, guid RunGuid) docker.CreateContainerOptions {
	namePrefix := runContext.NamespacePrefix("", string(guid))
	policyName := fmt.Sprintf("%vRP", namePrefix)
	restartPolicy := docker.RestartPolicy{
		Name:              policyName,
		MaximumRetryCount: 5,
	}
	anyPort := docker.PortBinding{HostPort: "0"}
	elbPort := docker.Port(fmt.Sprintf("%v/tcp", wa.BalancedInternalTCPPort))
	bindings := map[docker.Port][]docker.PortBinding{
		elbPort: []docker.PortBinding{anyPort}}
	hostConfig := &docker.HostConfig{
		PortBindings:  bindings,
		RestartPolicy: restartPolicy,
	}
	envsForDocker := EnvAsDockerKV(wa.Env())
	exposedPorts := map[docker.Port]struct{}{
		elbPort: struct{}{}}
	containerConfig := &docker.Config{
		Env:          envsForDocker,
		Image:        wa.Image,
		ExposedPorts: exposedPorts}
	return docker.CreateContainerOptions{Name: namePrefix, Config: containerConfig, HostConfig: hostConfig}
}

func (wa App) Env() map[string]string {
	envList := []map[string]string{}
	for _, env := range wa.AttachedEnvs {
		envList = append(envList, FindEnv(env).Env)
	}
	envList = append(envList, wa.ExtraEnv)
	return utils.FlattenHashes(envList...)
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
