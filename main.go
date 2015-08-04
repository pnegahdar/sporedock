package main

import (
	logging "github.com/op/go-logging"
	"github.com/pnegahdar/sporedock/grunts"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logging.SetLevel(logging.INFO, "main")
	gr := grunts.CreateAndRun("redis://localhost:6379", "testGroup", "myMachine", "127.0.0.1")
	gr.Wait()
}
