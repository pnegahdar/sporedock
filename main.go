package main

import (
	logging "github.com/op/go-logging"
	"github.com/pnegahdar/sporedock/grunts"
	"runtime"
	//	"fmt"
	//	"os"
	//	"os/signal"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logging.SetLevel(logging.INFO, "main")

	// Signal trap
	//	go func() {
	//		sigs := make(chan os.Signal, 1)
	//		signal.Notify(sigs, os.Interrupt)
	//		buf := make([]byte, 1<<20)
	//		<-sigs
	//		runtime.Stack(buf, true)
	//		fmt.Printf("REFERENCE TRACE:\n\n%s", buf)
	//		os.Exit(1)
	//	}()

	// RUN
	gr := grunts.CreateAndRun("redis://localhost:6379", "testGroup", "myMachine", "127.0.0.1", ":5000")
	//	go func(){
	//		<-time.After(time.Second * 4)
	//		gr.Stop()
	//	}()
	gr.Wait()
}
