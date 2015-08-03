package cluster

import (
	"github.com/pnegahdar/sporedock/utils"
	"github.com/samalba/dockerclient"
)

type WorkerApp struct {
	Count        int
	AttachedEnvs []string
	ExtraEnv     map[string]string
	Tags         map[string]string
	ID           string
	image        string
	Weight       float32
	Status       string
}

func (wa WorkerApp) HostConfig() dockerclient.HostConfig {
	return dockerclient.HostConfig{}
}
func (wa WorkerApp) ContainerConfig() dockerclient.ContainerConfig {
	envsForDocker := EnvAsDockerKV(wa.Env())
	return dockerclient.ContainerConfig{
		Env:   envsForDocker,
		Image: wa.image,
	}
}

func (wa WorkerApp) Env() map[string]string {
	envList := []map[string]string{}
	for _, env := range wa.AttachedEnvs {
		envList = append(envList, FindEnv(env).Env)
	}
	envList = append(envList, wa.ExtraEnv)
	return utils.FlattenHashes(envList...)
}

func (wa WorkerApp) Image() string {
	return wa.image
}

func (wa WorkerApp) Validate() error {
	return nil
}
