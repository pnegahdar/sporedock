package container

import (
	"fmt"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/utils"
	"github.com/samalba/dockerclient"
)

type WebApp struct {
	Count           int
	AttachedEnvs    []string
	ExtraEnv        map[string]string
	Tags            map[string]string
	ID              string
	Image           string
	Weight          float32
	BalancedTCPPort int
	Status          string
}

func (wa WebApp) RestartPolicy() dockerclient.RestartPolicy {
	policyName := fmt.Sprintf("SporedockRestartPolicy%vImage%v", wa.ID, wa.Image)
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
	bindings[fmt.Sprintf("%v/tcp", wa.BalancedTCPPort)] = []dockerclient.PortBinding{anyPort}
	return bindings
}

func (wa WebApp) Env() map[string]string {
	envList := []map[string]string{}
	for _, env := range wa.AttachedEnvs {
		envList = append(envList, cluster.FindEnv(env).Env)
	}
	return utils.FlattenHashes(wa.ExtraEnv, envList...)
}

func (wa WebApp) ContainerConfig() dockerclient.ContainerConfig {
	envsForDocker := cluster.EnvAsDockerKV(wa.Env())
	exposedPorts := map[string]struct{}{}
	exposedPorts[fmt.Sprintf("%v/tcp", wa.BalancedTCPPort)] = struct{}{}
	return dockerclient.ContainerConfig{
		Env:          envsForDocker,
		Image:        wa.Image,
		ExposedPorts: exposedPorts}
}
func (wa WebApp) Image() string {
	return wa.Image
}

func (wa WebApp) Identifier() string {
	return wa.ID
}

func (wa WebApp) TypeIdentifier() string {
	return "webapp"
}

func (wa WebApp) ToString() string {
	return utils.Marshall(wa)
}

func (wa WebApp) validate() error {
	return nil
}

func (wa WebApp) FromString(data string) (*WebApp, error) {
	wa := *WebApp{}
	utils.Unmarshall(data, wa)
	err := wa.validate()
	return wa, err
}
