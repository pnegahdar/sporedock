package cli

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/pnegahdar/sporedock/app"
	"os"
)

func StartMethod(c *cli.Context) {
	discovery := c.String(DiscoveryFlag.Name)
	app.StartServer(discovery)

}

func InitMethod(c *cli.Context) {
	discovery := c.String(DiscoveryFlag.Name)
	if discovery == "" {
		fmt.Printf("Discovery url required. Please pass with param: -%v\n", DiscoveryFlag.Name)
		os.Exit(1)
	}
	app.Initialize(discovery)
}
