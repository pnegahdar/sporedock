package cluster

import "github.com/samalba/dockerclient"

type ContainerApp interface {
	ContainerConfig() dockerclient.ContainerConfig
	HostConfig() dockerclient.HostConfig
	Image() string
	Env() map[string]string
	Identifier() string
}
