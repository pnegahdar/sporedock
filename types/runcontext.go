package types

import (
	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
	"net"
	"strings"
	"sync"
)

type RunContext struct {
	Config           *Config
	Store            SporeStore
	EventManager     *EventManager
	RPCManager       *RPCManager
	WebServerManager *WebServerManager
	DockerClient     *docker.Client
	CliManager       *CliManager
	initOnce         sync.Once
	sync.Mutex
}

func NewRunContext() *RunContext {
	return &RunContext{CliManager: NewCliManager()}
}

type Config struct {
	ConnectionString string
	MyGroup          string
	MyIP             net.IP
	MyType           SporeType
	MyMachineID      string
	WebServerBind    string
	RPCServerBind    string
	LoadBalancerBind string
}

func NewConfig(connectionString, myGroup, machineID string, myType SporeType, machineIP net.IP, webServerBind string, rpcServerBind string, loadBalacnerBind string) *Config {
	return &Config{
		ConnectionString: connectionString,
		MyType:           myType,
		MyGroup:          myGroup,
		MyMachineID:      machineID,
		MyIP:             machineIP,
		WebServerBind:    webServerBind,
		RPCServerBind:    rpcServerBind,
		LoadBalancerBind: loadBalacnerBind}
}

func NewConfigFromCli(cliContext *cli.Context) *Config {
	myName := RequiredStringArg(cliContext, FlagMyName.Name)
	myGroup := RequiredStringArg(cliContext, FlagGroupName.Name)
	myIP := net.ParseIP(RequiredStringArg(cliContext, FlagMyIP.Name))
	myType := SporeType(RequiredStringArg(cliContext, FlagMyType.Name))
	connectionString := RequiredStringArg(cliContext, FlagStoreConnectionString.Name)
	webserverBind := RequiredStringArg(cliContext, FlagWebServerBind.Name)
	rpcServerBind := RequiredStringArg(cliContext, FlagRPCServerBind.Name)
	loadBalancerBinder := RequiredStringArg(cliContext, FlagLoadBalancerBind.Name)
	return NewConfig(connectionString, myGroup, myName, myType, myIP, webserverBind, rpcServerBind, loadBalancerBinder)
}

func (rc RunContext) NamespacePrefixParts() []string {
	return []string{"sporedock", rc.Config.MyGroup}
}

func (rc RunContext) NamespacePrefix(joiner string, extra ...string) string {
	data := append(rc.NamespacePrefixParts(), extra...)
	return strings.Join(data, joiner)
}
