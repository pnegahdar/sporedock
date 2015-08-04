package utils

import "sync"

type SignalCast struct {
	mu        sync.Mutex
	listeners []chan bool
}

func (sc *SignalCast) Listen() chan bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	ret := make(chan bool)
	sc.listeners = append(sc.listeners, ret)
	return ret
}

func (sc *SignalCast) Signal() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	for _, listener := range sc.listeners {
		go func() {
			listener <- true
		}()
	}
}
