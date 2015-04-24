package main

import (
	logging "github.com/op/go-logging"
	"github.com/pnegahdar/sporedock/grunts"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logging.SetLevel(logging.INFO, "main")
	grunts.CreateAndRun()
}
