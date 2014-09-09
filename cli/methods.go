package cli

import (
	"fmt"
	"flag"
	"github.com/codegangsta/cli"
	"github.com/pnegahdar/SporeDock/worker"
	"github.com/pnegahdar/sporedock/settings"
	"os"
)

func StartMethod(c *cli.Context) {
	discovery := c.String(DiscoveryFlag.Name)
	fmt.Println(discovery)
	worker.Run(discovery)

}

func ConnectMethod(c *cli.Context){
	discovery := c.String(DiscoveryFlag.Name)
	if discovery == "" {
		fmt.Printf("Discovery url required. Please pass with param: -%v\n", DiscoveryFlag.Name)
		os.Exit(1)
	}
	f := &flag.Flag{Name : DiscoveryFlag.Name, Value: discovery}
	settings.AppConfig.Set("", f)
	fmt.Println("Configuration Initialized")
}
