package main

import (
	"fmt"
	logging "github.com/op/go-logging"
	"github.com/pnegahdar/sporedock/grunts"
	"github.com/pnegahdar/sporedock/utils"
	"os"
	"os/signal"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logging.SetLevel(logging.INFO, "main")

	gr := grunts.CreateAndRun("redis://localhost:6379", "testGroup", "myMachine", "127.0.0.1", ":5000")
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, os.Kill)
		<-sigs
		buf := make([]byte, 1<<20)
		runtime.Stack(buf, true)
		utils.LogInfo("Stopping grunts.")
		fmt.Println(string(buf))
		gr.Stop()
		os.Exit(1)
	}()
	gr.Wait()
	for {
	}
}
