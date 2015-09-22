package modules

import (
	"fmt"
	"github.com/gorilla/mux"
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
	initOnce   sync.Once
	stopCast   utils.SignalCast
	stopCastMu sync.Mutex
}

func (ws *WebServer) ShouldRun(runContext *types.RunContext) bool {
	return true
}

func (ws *WebServer) ProcName() string {
	return "WebServer"
}

func (ws *WebServer) Init(runContext *types.RunContext) {
	ws.initOnce.Do(func() {
		webServerRouter := mux.NewRouter().StrictSlash(true)
		webserverManager := &types.WebServerManager{WebServerBind: ":5000", WebServerRouter: webServerRouter}
		runContext.Lock()
		runContext.WebServerManager = webserverManager
		defer runContext.Unlock()
	})

}

func (ws *WebServer) Run(runContext *types.RunContext) {
	ws.stopCast = utils.SignalCast{}
	exit, _ := ws.stopCast.Listen()
	srv := &graceful.Server{
		Timeout: 1 * time.Second,
		Server:  &http.Server{Addr: runContext.WebServerManager.WebServerBind, Handler: runContext.WebServerManager.WebServerRouter},
	}
	go func(j *graceful.Server) {
		utils.LogInfo(fmt.Sprintf("Webserver started on %v", runContext.WebServerManager.WebServerBind))
		err := srv.ListenAndServe()
		if !strings.Contains(err.Error(), "use of closed network connection") {
			utils.HandleError(err)
		}
	}(srv)
	<-exit
	srv.Stop(srv.Timeout)
	utils.LogInfo(fmt.Sprintf("Webserver stopped on %v", runContext.WebServerManager.WebServerBind))
}

func (ws *WebServer) Stop() {
	ws.stopCast.Signal()
}
