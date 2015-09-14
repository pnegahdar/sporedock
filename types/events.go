package types

import (
	"github.com/pnegahdar/sporedock/utils"
	"sync"
)

type Event string

var EventAll Event = "*"
var EventDockerAppStart Event = "docker:app:started"

type EventMessage struct {
	Emitter   SporeID
	EmitterIP string
	Event     Event
}

func (ev *Event) emit(rc *RunContext, channels ...string) {
	myID := rc.MyMachineID
	message := EventMessage{Emitter: SporeID(myID), EmitterIP: rc.MyIP.String(), Event: *ev}
	err := rc.Store.Publish(message, channels...)
	utils.HandleError(err)
}

func (ev *Event) EmitAll(rc *RunContext) {
	spores, err := AllSpores(rc)
	utils.HandleError(err)
	sporeIDS := []string{}
	for _, spore := range spores {
		sporeIDS = append(sporeIDS, spore.ID)
	}
	ev.emit(rc, sporeIDS...)
}

func (ev *Event) EmitToSelf(rc *RunContext) {
	ev.emit(rc, rc.MyMachineID)
}

func (ev *Event) Matches(matching Event) bool {
	if *ev == matching || *ev == EventAll {
		return true
	}
	return false
}

type EventManager struct {
	listeners  map[Event]map[string]chan EventMessage
	manager    *SubscriptionManager
	initOnce   sync.Once
	ExitSignal utils.SignalCast
	sync.Mutex
}

func (em *EventManager) init(rc *RunContext) {
	em.listeners = map[Event]map[string]chan EventMessage{}

}

func (em *EventManager) BroadcastToListeners(message EventMessage) {
	em.Lock()
	if listeners, ok := em.listeners[message.Event]; ok {
		for _, listener := range listeners {
			listener := listener
			go func() { select {
				case listener <- message:
				default:
					return
				}
			}()
		}
	}
	em.Unlock()
}

func (em *EventManager) Listen(rc *RunContext, event Event, exit *utils.SignalCast) chan EventMessage {
	em.initOnce.Do(func() { em.init(rc) })
	message := make(chan EventMessage)
	listenerID := utils.GenGuid()
	em.Lock()
	if _, ok := em.listeners[event]; !ok {
		em.listeners[event] = map[string]chan EventMessage{}
	}
	em.listeners[event][listenerID] = message
	em.Unlock()
	go func() {
		exitFromParent, _ := em.ExitSignal.Listen()
		exitFromChild, _ := exit.Listen()
		removeMe := func() {
			em.Lock()
			close(em.listeners[event][listenerID])
			delete(em.listeners[event], listenerID)
			em.Unlock()
		}
		for {
			select {
			case <-exitFromParent:
				removeMe()
				return
			case <-exitFromChild:
				removeMe()
				return
			}
		}
	}()
	return message
}
