package cli

import (
	"github.com/codegangsta/cli"
	"github.com/pnegahdar/sporedock/settings"
)

var CliApp = cli.NewApp()

func init() {
	CliApp.Name = settings.AppName
	CliApp.Usage = settings.AppUsage
	CliApp.Version = settings.AppVersion
	CliApp.Author = settings.AppAuthor
	CliApp.Commands = []cli.Command{StartCommand, ConnectCommand}
}
