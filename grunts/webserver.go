package grunts

import (
	"fmt"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"gopkg.in/tylerb/graceful.v1"
	"net/http"
	"strings"
	"sync"
	"time"
)

type WebServer struct {
	sync.Mutex
	stopCast   utils.SignalCast
	stopCastMu sync.Mutex
}

func (ws WebServer) ShouldRun(runContext *types.RunContext) bool {
	return true
}

func (ws WebServer) ProcName() string {
	return "WebServer"
}

func (ws *WebServer) Run(runContext *types.RunContext) {
	ws.Lock()
	defer ws.Unlock()
	ws.stopCast = utils.SignalCast{}
	exit, _ := ws.stopCast.Listen()
	srv := &graceful.Server{
		Timeout: 1 * time.Second,
		Server:  &http.Server{Addr: runContext.WebServerBind, Handler: runContext.WebServerRouter},
	}
	go func(j *graceful.Server) {
		utils.LogInfo(fmt.Sprintf("Webserver started on %v", runContext.WebServerBind))
		err := srv.ListenAndServe()
		if !strings.Contains(err.Error(), "use of closed network connection") {
			utils.HandleError(err)
		}
	}(srv)
	<-exit
	srv.Stop(srv.Timeout)
	utils.LogInfo(fmt.Sprintf("Webserver stopped on %v", runContext.WebServerBind))
}

func (ws *WebServer) Stop() {
	ws.stopCastMu.Lock()
	defer ws.stopCastMu.Unlock()
	ws.stopCast.Signal()
}
