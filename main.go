package main

import (
	logging "github.com/op/go-logging"
	"github.com/pnegahdar/sporedock/cli"
	"github.com/pnegahdar/sporedock/config"
	"os"
)

func main() {
	logging.SetLevel(logging.DEBUG, "main")
	config.ImportClusterConfigFromFile("sample_cluster.json")
	cli.CliApp.Run(os.Args)
}
