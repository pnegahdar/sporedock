package modules

import (
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
)

type Event string

var EventDockerAppStart Event = "docker:app:started"

type EventMessage struct {
	Emitter   cluster.SporeID
	EmitterIP string
	Event     Event
}

func (ev Event) emit(rc *types.RunContext, channels ...string) {
	myID := rc.MyMachineID
	message := EventMessage{Emitter: cluster.SporeID(myID), EmitterIP: rc.MyIP.String(), Event: ev}
	rc.Store.Publish(message, channels...)
}

func (ev Event) EmitAll(rc *types.RunContext) {
	spores, err := cluster.AllSpores(rc)
	utils.HandleError(err)
	sporeIDS := []string{}
	for _, spore := range (spores) {
		sporeIDS = append(sporeIDS, spore.ID)
	}
}


type EventManager struct {
	listeners map[string]chan EventMessage
}

func (em *EventManager) Listen(channel string) chan EventMessage {
	return make(chan EventMessage)
}

func (em *EventManager) ListenMe(rc *types.RunContext) chan EventMessage{
	return em.Listen(rc.MyMachineID)
}
