package main

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/pnegahdar/sporedock/registry"
	"github.com/pnegahdar/sporedock/utils"
	"os"
	"os/signal"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logging.SetLevel(logging.INFO, "main")

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, os.Kill)
		<-sigs
		buf := make([]byte, 1<<20)
		runtime.Stack(buf, true)
		utils.LogInfo("Stopping modules.")
		fmt.Println(string(buf))
		os.Exit(1)
	}()
	go func() {
		for {
			utils.LogInfoF("Goroutine Count: %v", runtime.NumGoroutine())
			<-time.After(time.Second * 60)
		}
	}()
	moduleRegister := registry.Create()
	moduleRegister.RunCli()
}
