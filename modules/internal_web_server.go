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
		webserverManager := &types.WebServerManager{WebServerRouter: webServerRouter}
		runContext.Lock()
		defer runContext.Unlock()
		runContext.WebServerManager = webserverManager
	})

}

func (ws *WebServer) Run(runContext *types.RunContext) {
	ws.stopCast = utils.SignalCast{}
	exit, _ := ws.stopCast.Listen()
	srv := &graceful.Server{
		Timeout: 1 * time.Second,
		Server:  &http.Server{Addr: runContext.Config.WebServerBind, Handler: runContext.WebServerManager.WebServerRouter},
	}
	go func(j *graceful.Server) {
		utils.LogInfo(fmt.Sprintf("Webserver started on %v", runContext.Config.WebServerBind))
		err := srv.ListenAndServe()
		if !strings.Contains(err.Error(), "use of closed network connection") {
			utils.HandleError(err)
		}
	}(srv)
	<-exit
	srv.Stop(srv.Timeout)
	utils.LogInfo(fmt.Sprintf("Webserver stopped on %v", runContext.Config.WebServerBind))
}

func (ws *WebServer) Stop() {
	ws.stopCast.Signal()
}
