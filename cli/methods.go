package cli

import (
	"os"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/pnegahdar/SporeDock/worker"

)


func StartMethod(c *cli.Context){
	discovery := c.String(DiscoveryFlag.Name)
	if(discovery == ""){
		fmt.Printf("Discovery url required. Please pass with param: -%v\n", DiscoveryFlag.Name)
		os.Exit(1)
	}
	worker.Run(discovery)

}

