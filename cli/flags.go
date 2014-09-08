package cli

import (
	"github.com/codegangsta/cli"
)

var DiscoveryFlag = cli.StringFlag{
	Name: "discovery",
	Usage: "The Etcd discovery URL (required)",
}
