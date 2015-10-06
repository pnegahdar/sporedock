package types

import (
	"github.com/codegangsta/cli"
	"github.com/pnegahdar/sporedock/utils"
)

type CliManager struct {
	Cli *cli.App
}

func (clm *CliManager) AddCommand(comms ...cli.Command) {
	clm.Cli.Commands = append(clm.Cli.Commands, comms...)
}

func RequiredStringArg(c *cli.Context, flagName string) string {
	result := c.String(flagName)
	if result == "" {
		utils.LogErrorF("Flag must be set: %v", flagName)
	}
	return result
}

func NewCliManager() *CliManager {
	cliApp := cli.NewApp()
	cliApp.Name = "SporeDock"
	cliApp.Usage = "Kill it in the cloud"
	cliApp.Version = "Pre-Greek-Alphabet"
	cliApp.Author = "Parham Negahdar <pnegahdar@gmail.com>"
	return &CliManager{Cli: cliApp}
}

var FlagStoreConnectionString = cli.StringFlag{
	Name:   "connection-string",
	Value:  "redis://127.0.0.1:6379",
	Usage:  "The approprate URI for the store connection",
	EnvVar: "SPOREDOCK_CONNECTION_STRING"}

var FlagGroupName = cli.StringFlag{
	Name:   "my-group",
	Usage:  "The group namespace for this sporedock cluster",
	EnvVar: "SPOREDOCK_GROUP"}

var FlagMyName = cli.StringFlag{
	Name:   "my-name",
	Usage:  "A unique identifier for this node in the cluster",
	EnvVar: "SPOREDOCK_MY_NAME"}

var FlagMyType = cli.StringFlag{
	Name:   "my-type",
	Value:  "worker",
	Usage:  "The Type of member",
	EnvVar: "SPOREDOCK_MY_NAME"}

var FlagMyIP = cli.StringFlag{
	Name:   "my-ip",
	Usage:  "The IP for this box reachable by other nodes in the cluseter",
	EnvVar: "SPOREDOCK_MY_IP"}

var FlagWebServerBind = cli.StringFlag{
	Name:   "webserver-bind",
	Value:  ":5000",
	Usage:  "Addr to bind webserver on",
	EnvVar: "SPOREDOCK_WEBSERVER_BIND"}

var FlagRPCServerBind = cli.StringFlag{
	Name:   "rpcserver-bind",
	Value:  ":5001",
	Usage:  "Addr to bind rpc server on",
	EnvVar: "SPOREDOCK_RPCSERVER_BIND"}

var FlagLoadBalancerBind = cli.StringFlag{
	Name:   "loadbalancer-bind",
	Value:  ":8008",
	Usage:  "Addr to bind loadbalancer on",
	EnvVar: "SPOREDOCK_LOADBALANCER_BIND"}

var RunCommandFlags = []cli.Flag{FlagStoreConnectionString, FlagGroupName, FlagMyIP, FlagMyName, FlagWebServerBind, FlagRPCServerBind, FlagLoadBalancerBind, FlagMyType}
