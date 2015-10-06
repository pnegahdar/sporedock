package types

import (
	"github.com/gorilla/mux"
	"sync"
)

type WebServerManager struct {
	WebServerRouter *mux.Router
	sync.Mutex
}
