package types

import "github.com/codegangsta/cli"

type CliManager struct {
	Cli *cli.App
}

func (clm *CliManager) AddCommand(comm cli.Command) {
	if len(clm.Cli.Commands) == 0 {
		clm.Cli.Commands = []cli.Command{comm}
	} else {
		clm.Cli.Commands = append(clm.Cli.Commands, comm)
	}
}

func NewCliManager() *CliManager {
	cliApp := cli.NewApp()
	cliApp.Name = "SporeDock"
	cliApp.Usage = "Kill it in the cloud"
	cliApp.Version = "Pre-Greek-Alphabet"
	cliApp.Author = "Parham Negahdar <pnegahdar@gmail.com>"
	return &CliManager{Cli: cliApp}
}
