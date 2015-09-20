package types

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/pnegahdar/sporedock/utils"
)

type AppID string
type App struct {
	Count                   int
	CountRemaining          int `json:"-"`
	Scheduler               string
	PinSpore                string
	AttachedEnvs            []string
	ExtraEnv                map[string]string
	Tags                    map[string]string
	ID                      AppID
	Image                   string
	BalancedInternalTCPPort int
	Sizable
}

func (a App) Size() float64 {
	return GetSize(a.Cpus, a.Mem)
}

type Apps []App

func (a Apps) Len() int      { return len(a) }
func (a Apps) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Apps) Less(i, j int) bool {
	return GetSize(a[i].Cpus, a[i].Mem) < GetSize(a[j].Cpus, a[j].Mem)
}

func DockerPort(interalPort int) docker.Port {
	return docker.Port(fmt.Sprintf("%v/tcp", interalPort))
}

func (wa App) DockerContainerOptions(runContext *RunContext, guid RunGuid) docker.CreateContainerOptions {
	namePrefix := dockerNameSpacePrefix(runContext, string(wa.ID), string(guid))
	policyName := fmt.Sprintf("%vRP", namePrefix)
	restartPolicy := docker.RestartPolicy{
		Name:              policyName,
		MaximumRetryCount: 5,
	}
	anyPort := docker.PortBinding{HostPort: "0"}
	elbPort := DockerPort(wa.BalancedInternalTCPPort)
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

func (wa *App) Validate(rc *RunContext) error {
	return nil
}

func (wa *App) GetID() string {
	return string(wa.ID)
}

func AllApps(runContext *RunContext) ([]App, error) {
	apps := []App{}
	err := runContext.Store.GetAll(&apps, 0, SentinelEnd)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func GetPortOn(runContext *RunContext, spore *Spore, app App, runGuid RunGuid) string {
	containersRunning, err := runContext.DockerClient.ListContainers(docker.ListContainersOptions{All: false})
	utils.HandleError(err)
	appName := fullDockerAppName(runGuid, containersRunning)
	if appName == "" {
		return "0"
	}
	resp, err := runContext.DockerClient.InspectContainer(appName)
	if err != nil {
		utils.LogWarnF("Had issue finding container %v", appName)
		return "0"
	}
	binding, ok := resp.NetworkSettings.Ports[DockerPort(app.BalancedInternalTCPPort)]
	if !ok {
		return "0"
	}
	return binding[0].HostPort
}
