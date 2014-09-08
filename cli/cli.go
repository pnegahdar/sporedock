package cli

import (
	"github.com/codegangsta/cli"
)

var CliApp = cli.NewApp()


func init() {
	CliApp.Name = "SporeDocker"
	CliApp.Usage = "Docker discovery based distribution and load balanced"
	CliApp.Version = "1.0.0"
	CliApp.Author = "Parham Negahdar"
	CliApp.Commands = []cli.Command{StartCommand,}
}
