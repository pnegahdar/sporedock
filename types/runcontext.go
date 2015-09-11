package types

import (
	"github.com/fsouza/go-dockerclient"
	"net"
	"strings"
	"sync"
)

type RunContext struct {
	Store            SporeStore
	MyMachineID      string
	MyIP             net.IP
	MyType           SporeType
	MyGroup          string
	EvetnManager     *EventManager
	RPCManager       *RPCManager
	WebServerManager *WebServerManager
	DockerClient     *docker.Client
	initOnce         sync.Once
	sync.Mutex
}

func (rc RunContext) NamespacePrefixParts() []string {
	return []string{"sporedock", rc.MyGroup}
}

func (rc RunContext) NamespacePrefix(joiner string, extra ...string) string {
	data := append(rc.NamespacePrefixParts(), extra...)
	return strings.Join(data, joiner)
}
