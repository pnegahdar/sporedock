package main

import (
	"fmt"
	"github.com/pnegahdar/sporedock/utils"
	"runtime"
	"github.com/op/go-logging"
	"os"
	"os/signal"
	"github.com/pnegahdar/sporedock/sporedock"
)


func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logging.SetLevel(logging.INFO, "main")

	gr := sporedock.CreateAndRun("redis://localhost:6379", "testGroup", "myMachine", "127.0.0.1", ":5000", ":5001")
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, os.Kill)
		<-sigs
		buf := make([]byte, 1<<20)
		runtime.Stack(buf, true)
		utils.LogInfo("Stopping modules.")
		fmt.Println(string(buf))
		gr.Stop()
		os.Exit(1)
	}()
	gr.Wait()
	for {
	}
}
