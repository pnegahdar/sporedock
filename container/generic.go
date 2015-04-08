package container

import "github.com/samalba/dockerclient"

type ContainerApp interface {
	ContainerConfig() dockerclient.ContainerConfig
	HostConfig() dockerclient.HostConfig
	Image() string
	Name() string
}
