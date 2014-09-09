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

var InitCommand = cli.Command{
	Name:      "init",
	ShortName: "i",
	Usage:     "Set the config discovery url",
	Action:    InitMethod,
	Flags:     []cli.Flag{DiscoveryFlag},
}
