package modules

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

var loadBalancerDebounceInterval = time.Second * 5

type LoadBalancerModule struct {
	initOnce   sync.Once
	stopCast   utils.SignalCast
	runContext *types.RunContext
}

func (lbm *LoadBalancerModule) Init(runContext *types.RunContext) {
	lbm.initOnce.Do(func() {
		lbm.runContext = runContext
	})
}

func (lbm *LoadBalancerModule) ProcName() string {
	return "LoadBalancer"
}

func (lbm *LoadBalancerModule) Stop() {
	lbm.stopCast.Signal()
}

func (lbm *LoadBalancerModule) Run(runContext *types.RunContext) {
	loadBalancer := &types.LoadBalancer{}
	httpServer := &http.Server{
		Addr:    runContext.Config.LoadBalancerBind,
		Handler: loadBalancer,
	}
	srv := &graceful.Server{
		Timeout: 1 * time.Second,
		Server:  httpServer,
	}
	go func(j *graceful.Server) {
		utils.LogInfo(fmt.Sprintf("Loadbalancer started on %v", runContext.Config.LoadBalancerBind))
		err := srv.ListenAndServe()
		if !strings.Contains(err.Error(), "use of closed network connection") {
			utils.HandleError(err)
		}
	}(srv)
	appHostMeta, err := types.NewMeta(&types.AppHost{})
	utils.HandleError(err)
	appHostEvent := types.StoreEvent(types.StorageActionAll, appHostMeta)
	eventFeed := runContext.EventManager.ListenDebounced(runContext, &lbm.stopCast, loadBalancerDebounceInterval, types.EventDockerAppStart, appHostEvent)
	exit, _ := lbm.stopCast.Listen()
	loadBalancer.Update(runContext)
	for {
		select {
		case <-eventFeed:
			loadBalancer.Update(runContext)
		case <-exit:
			srv.Stop(srv.Timeout)
			utils.LogInfo(fmt.Sprintf("LoadBalancer stopped on %v", runContext.Config.LoadBalancerBind))
			return
		}
	}
}

func (lb *LoadBalancerModule) ShouldRun(runContext *types.RunContext) bool {
	return true
}
