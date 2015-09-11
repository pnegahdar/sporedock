package modules

import (
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sync"
)

type EventModule struct {
	initOnce sync.Once
	stopCast utils.SignalCast
}

func (em *EventModule) Init(rc *types.RunContext) {
	rc.Lock()
	defer rc.Unlock()
	rc.EvetnManager = &types.EventManager{}
}

func (em *EventModule) Run(rc *types.RunContext) {
	exit, _ := em.stopCast.Listen()
	<-exit
}

func (em *EventModule) ProcName() string {
	return "Events"
}

func (em *EventModule) ShouldRun(rc *types.RunContext) bool {
	return true
}

func (em *EventModule) Stop(rc *types.RunContext) {
	em.stopCast.Signal()
}
