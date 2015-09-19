package modules

import (
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sync"
)

type Cli struct {
	initOnce   sync.Once
	stopCast   utils.SignalCast
	runContext *types.RunContext
}

func (module *Cli) Init(runContext *types.RunContext) {
	module.initOnce.Do(func() {
		module.runContext = runContext
		runContext.Lock()
		runContext.CliManager = types.NewCliManager()
		runContext.Unlock()
	})
}

func (module *Cli) ProcName() string {
	return "Cli"
}

func (module *Cli) Stop() {
	module.stopCast.Signal()
}

func (module *Cli) Run(runContext *types.RunContext) {
	exit, _ := module.stopCast.Listen()
	<-exit
}

func (module *Cli) ShouldRun(runContext *types.RunContext) bool {
	return true
}
