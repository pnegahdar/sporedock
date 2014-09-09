package cli

import (
	"github.com/codegangsta/cli"
)

var StartCommand = cli.Command{
	Name:      "start",
	ShortName: "s",
	Usage:     "Start the SporeDock server on this node.",
	Action:    StartMethod,
}


var ConnectCommand = cli.Command{
	Name: "connect",
	ShortName: "c",
	Usage: "Set the config discovery url",
	Action: ConnectMethod,
	Flags:     []cli.Flag{DiscoveryFlag},
}
