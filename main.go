package main

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/pnegahdar/sporedock/sporedock"
	"github.com/pnegahdar/sporedock/utils"
	"os"
	"os/signal"
	"runtime"
	"time"
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
	for {
		<-time.After(time.Second * 5)
		fmt.Println(runtime.NumGoroutine())
	}
}
