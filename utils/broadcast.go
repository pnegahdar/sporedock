package utils

import (
	"errors"
	"sync"
)

type SignalCast struct {
	sync.Mutex
	name           string
	stop           chan bool
	alreadyFlipped bool
	listeners      map[string]chan bool
}

func (sc *SignalCast) Listen(name string) chan bool {
	sc.Lock()
	defer sc.Unlock()
	sc.init()
	ret := make(chan bool, 1)
	if sc.alreadyFlipped {
		ret <- true
	}
	_, ok := sc.listeners[name]
	if !ok {
		sc.listeners[name] = ret
	} else {
		HandleError(errors.New("Duplicate handler added" + sc.name))
	}
	return ret
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
	sc.Lock()
	defer sc.Unlock()
	sc.init()
	wg := sync.WaitGroup{}
	sc.alreadyFlipped = true
	for k := range sc.listeners {
		wg.Add(1)
		go func(j string) {
			sc.listeners[j] <- true
			wg.Done()
		}(k)
	}
	wg.Wait()
}
