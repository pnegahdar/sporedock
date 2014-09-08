package cli

import(
	"github.com/codegangsta/cli"
)
var StartCommand = cli.Command{
Name:      "start",
ShortName: "s",
Usage:     "Start the SporeDock server on this node.",
Action: 	StartMethod,
Flags: 		[]cli.Flag{DiscoveryFlag},
}
