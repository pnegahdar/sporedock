package main

import (
	logging "github.com/op/go-logging"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/grunts"
	"github.com/pnegahdar/sporedock/store"
	"net"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logging.SetLevel(logging.INFO, "main")
	connectionString := "redis://localhost:6379"
	groupName := "testGroup"
	machineID := "myMachine"
	myIP := net.ParseIP("127.0.0.1")
	myType := cluster.TypeSporeMember

	// Initialize workers
	genericWorker := grunts.TestRunner{}
	store := store.CreateStore(connectionString, groupName)

	// Create Run Context
	runContext := grunts.RunContext{myMachineID: machineID, store: store, myIP: myIP, myType: myType, myGroup: groupName}

	// Register and run
	gruntRegistry := grunts.GruntRegistry{Context: runContext}
	gruntRegistry.Start(genericWorker, store)

}
