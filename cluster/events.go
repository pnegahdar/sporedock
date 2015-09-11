package cluster

import (
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sync"
)

type Event string

var EventAll Event = ""
var EventDockerAppStart Event = "docker:app:started"

type EventMessage struct {
	Emitter   SporeID
	EmitterIP string
	Event     Event
}

func (ev Event) emit(rc *types.RunContext, channels ...string) {
	myID := rc.MyMachineID
	message := EventMessage{Emitter: SporeID(myID), EmitterIP: rc.MyIP.String(), Event: ev}
	rc.Store.Publish(message, channels...)
}

func (ev Event) EmitAll(rc *types.RunContext) {
	spores, err := AllSpores(rc)
	utils.HandleError(err)
	sporeIDS := []string{}
	for _, spore := range (spores) {
		sporeIDS = append(sporeIDS, spore.ID)
	}
}

func (ev *Event) Matches(matching Event) bool {
	if ev == matching || ev == EventAll {
		return true
	}
	return false
}

type eventListner struct {
	exitChan *utils.SignalCast
	receive  chan EventMessage
}


type EventManager struct {
	listeners map[Event][]eventListner
	manager   *types.SubscriptionManager
	initOnce  sync.Once
}

func (em *EventManager) init(rc types.RunContext) {
	em.listeners = map[Event][]eventListner{}

}

func (em *EventManager) Listen(rc *types.RunContext, event Event, exit *utils.SignalCast) chan EventMessage {
	sync.Once.Do(func() {em.init(rc)})
	go func() {
		exitChan, _ := exit.Listen()
		for {
			select {
			case <-exitChan:
				return
			case <-
			}
		}
	}()
}

