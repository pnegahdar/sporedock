package container

import (
    "github.com/samalba/dockerclient"
    "github.com/pnegahdar/sporedock/cluster"
    "github.com/pnegahdar/sporedock/utils"
    "github.com/pnegahdar/sporedock/store"
)

type WorkerApp struct {
    Count  int
    AttachedEnvs    []string
    ExtraEnv map[string]string
    Tags   map[string]string
    ID     string
    Image  string
    Weight float32
}

func (wa WorkerApp) HostConfig() dockerclient.HostConfig {
    return dockerclient.HostConfig{}
}
func (wa WorkerApp) ContainerConfig() dockerclient.ContainerConfig {
    envsForDocker := cluster.EnvAsDockerKV(wa.Env())
    return dockerclient.ContainerConfig{
        Env: envsForDocker,
        Image: wa.Image,
    }
}

func (wa WorkerApp) Env() map[string]string {
    envList := []map[string]string{}
    for _, env := range (wa.AttachedEnvs) {
        envList = append(envList, cluster.FindEnv(env).Env)
    }
    return utils.FlattenHashes(wa.ExtraEnv, envList...)
}

func (wa WorkerApp) Image() string {
    return wa.Image
}

func (wa WorkerApp) Identifier() string {
    return wa.ID
}

func (wa WorkerApp) TypeIdentifier() string {
    return "workerapp"
}

func (wa WorkerApp) ToString() string {
    return utils.Marshall(wa)
}

func (wa WorkerApp) validate() error {
    return nil
}

func (wa *WorkerApp) FromString(data string) (*store.Storable, error) {
    utils.Unmarshall(data, wa)
    err := wa.validate()
    return wa, err
}

