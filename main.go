package main

import (
	logging "github.com/op/go-logging"
	"runtime"
    "github.com/pnegahdar/sporedock/grunts"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logging.SetLevel(logging.INFO, "main")

    // Create Run Context
    runContext := grunts.RunContext{}
    gruntRegistry := grunts.GruntRegistry{Context: runContext}

    // Create and Register Grunts
    genericWorker := grunts.TestRunner{}
    gruntRegistry.Start(genericWorker)

}
