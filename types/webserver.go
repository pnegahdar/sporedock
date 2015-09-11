package types

import (
	"github.com/gorilla/mux"
	"sync"
)

type WebServerManager struct {
	WebServerBind   string
	WebServerRouter *mux.Router
	sync.Mutex
}
