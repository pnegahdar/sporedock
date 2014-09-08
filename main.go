package main

import (
	"os"

	"github.com/pnegahdar/SporeDock/cli"
)

func main() {
	cli.CliApp.Run(os.Args)
}
