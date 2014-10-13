package main

import (
	"fmt"
	"github.com/pnegahdar/sporedock/cli"
	"github.com/pnegahdar/sporedock/manifest"
	"os"
)

func app() {
	cli.CliApp.Run(os.Args)
}

func main() {
	fmt.Printf("%v\n" , manifest.ImportClusterConfigFromFile("sample_cluster.json"))
}
