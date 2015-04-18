package cluster

import (
    "fmt"
    "github.com/pnegahdar/sporedock/utils"
    "github.com/samalba/dockerclient"
    "github.com/pnegahdar/sporedock/types"
)

type WebApp struct {
    Count           int
    AttachedEnvs    []string
    ExtraEnv        map[string]string
    Tags            map[string]string
    ID              string
    image           string
    Weight          float32
    BalancedTCPPort int
    Status          string
}

func (wa WebApp) RestartPolicy() dockerclient.RestartPolicy {
    policyName := fmt.Sprintf("SporedockRestartPolicy%vImage%v", wa.ID, wa.image)
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
        envList = append(envList, FindEnv(env).Env)
    }
    envList = append(envList, wa.ExtraEnv)
    return utils.FlattenHashes(envList...)
}

func (wa WebApp) ContainerConfig() dockerclient.ContainerConfig {
    envsForDocker := EnvAsDockerKV(wa.Env())
    exposedPorts := map[string]struct {}{}
    exposedPorts[fmt.Sprintf("%v/tcp", wa.BalancedTCPPort)] = struct {}{}
    return dockerclient.ContainerConfig{
        Env:          envsForDocker,
        Image:        wa.image,
        ExposedPorts: exposedPorts}
}
func (wa WebApp) Image() string {
    return wa.image
}

func (wa WebApp) Identifier() string {
    return wa.ID
}

func (wa WebApp) TypeIdentifier() string {
    return "webapp"
}

func (wa WebApp) ToString() string {
    resp, err := utils.Marshall(wa)
    utils.HandleError(err)
    return resp
}

func (wa WebApp) validate() error {
    return nil
}

func (wa WebApp) FromString(data string) (types.Storable, error) {
    wa = WebApp{}
    utils.Unmarshall(data, wa)
    err := wa.validate()
    return wa, err
}

func GetAllWebApps(rc *types.RunContext) []WebApp{
    retType := WebApp{}
    webapps := []WebApp{}
    storables := rc.Store.GetAll(retType)
    for _, storable := range(storables){
        webapps = append(webapps, storable.(WebApp))
    }
    return webapps
}

