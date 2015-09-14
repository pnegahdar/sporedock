package utils

import (
	"errors"
	"sync"
)

type SignalCast struct {
	mu             sync.Mutex
	name           string
	stop           chan bool
	alreadyFlipped bool
	listeners      map[string]chan bool
}

func (sc *SignalCast) Listen() (chan bool, string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.init()
	ret := make(chan bool)
	if sc.alreadyFlipped {
		go func() {
			select {
			case ret <- true:
			default:
				return
			}}()
	}
	name := GenGuid()
	if _, ok := sc.listeners[name]; !ok {
		sc.listeners[name] = ret
	} else {
		HandleError(errors.New("Duplicate handler added" + sc.name))
	}
	return ret, name
}

func (sc *SignalCast) init() {
	if sc.stop == nil {
		sc.stop = make(chan bool)
	}
	if sc.listeners == nil {
		sc.listeners = make(map[string]chan bool)
	}
	if sc.name == "" {
		sc.name = GenGuid()
	}
}

func (sc *SignalCast) Signal() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.init()
	sc.alreadyFlipped = true
	for k := range sc.listeners {
		k := k
		go func(j string) {
			select {
			case sc.listeners[j] <- true:
			default:
				return
			}
		}(k)
	}
}
