package grunt

import "github.com/pnegahdar/sporedock/store"

type RunContext struct {


}

type GruntWorker interface{
    Name() string
    Run(runContext RunContext) chan bool
    ShouldRun(mySpore store.Spore) bool

}
