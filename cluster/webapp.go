package cluster

import (
	"errors"
	"fmt"
	"github.com/pnegahdar/sporedock/types"
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
	BalancedTCPPort int
}

var (
	ErrIDEmpty = errors.New("Webapp ID cannot be empty.")
	ErrIDExists = errors.New("Webapp with that ID already exists please delete and try again.")
)

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
		Image:        wa.Image,
		ExposedPorts: exposedPorts}
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

func (wa WebApp) validate(rc *types.RunContext) error {
	if wa.ID == "" { // Todo(parham): address race condition
		return ErrIDEmpty
	}
	webapp, err := GetWebapp(rc, wa.ID)
	if types.RewrapError(err) != types.ErrEmptyQuery {
		if webapp.ID == wa.ID {
			return ErrIDExists
		}
	}
	return err
}

func NewWebApp(id string, image string, balancedTcpPort int) *WebApp {
	return &WebApp{
		ID: id,
		Image: image,
		BalancedTCPPort: balancedTcpPort,
		Count: 1,
	}
}

func (wa WebApp) FromString(data string, rc *types.RunContext) (types.Storable, error) {
	wa = WebApp{}
	utils.Unmarshall(data, &wa)
	err := wa.validate(rc)
	return wa, err
}

func GetAllWebApps(rc *types.RunContext) ([]WebApp, error) {
	retType := WebApp{}
	webapps := []WebApp{}
	storables, err := rc.Store.GetAll(retType)
	if err != nil {
		return webapps, err
	}
	for _, storable := range storables {
		webapps = append(webapps, storable.(WebApp))
	}
	return webapps, nil
}

func GetWebapp(rc *types.RunContext, id string) (WebApp, error) {
	ret := WebApp{ID: id}
	webapp, err := rc.Store.Get(ret)
	if err != nil {
		return ret, err
	}
	return webapp.(WebApp), nil

}
