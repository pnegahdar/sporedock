package modules

import (
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sync"
)

type EventModule struct {
	initOnce   sync.Once
	stopCast   utils.SignalCast
	subManager *types.SubscriptionManager
}

func (em *EventModule) Init(rc *types.RunContext) {
	rc.Lock()
	defer rc.Unlock()
	rc.EventManager = &types.EventManager{}
	subManager, err := rc.Store.Subscribe(rc.Config.MyMachineID)
	utils.HandleError(err)
	em.subManager = subManager
}

func (em *EventModule) Run(rc *types.RunContext) {
	go func() {
		exit, _ := em.stopCast.Listen()
		for {
			select {
			case message := <-em.subManager.Messages:
				eventMessage := &types.EventMessage{}
				utils.Unmarshall(message, eventMessage)
				rc.EventManager.BroadcastToListeners(*eventMessage)
			case <-exit:
				rc.EventManager.ExitSignal.Signal()
				em.subManager.Exit.Signal()
				return
			}
		}
	}()
	exit, _ := em.stopCast.Listen()
	<-exit
}

func (em *EventModule) ProcName() string {
	return "Events"
}

func (em *EventModule) ShouldRun(rc *types.RunContext) bool {
	return true
}

func (em *EventModule) Stop() {
	em.stopCast.Signal()
}
