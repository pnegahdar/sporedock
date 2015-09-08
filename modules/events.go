package modules

import (
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/types"
)

type Event struct {
	Body string
	Tags []string
}

type EventMessage struct {
	Emitter   cluster.SporeID
	EmitterIP string
	Event     Event
}

func (ev Event) Emit(rc *types.RunContext) {
}
