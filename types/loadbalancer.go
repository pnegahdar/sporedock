package types

import (
	"fmt"
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/roundrobin"
	"github.com/mailgun/oxy/stream"
	"github.com/pnegahdar/sporedock/utils"
	"net/http"
	"net/url"
	"sync"
)

type streamRR struct {
	stream *stream.Streamer
	rr     *roundrobin.RoundRobin
}

type HostHandlers map[Hostname]streamRR

type LoadBalancer struct {
	sync.RWMutex
	Handlers HostHandlers
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	host := req.Host
	lb.RLock()
	handler, ok := lb.Handlers[Hostname(host)]
	lb.RUnlock()
	if !ok {
		http.NotFound(w, req)
	} else {
		handler.stream.ServeHTTP(w, req)
	}
}

func (lb *LoadBalancer) Update(runContext *RunContext) {
	lb.Lock()
	defer lb.Unlock()
	utils.LogInfoF("Running LB update")
	lb.Handlers = GetMyHostMap(runContext)
}

func GetMyHostMap(runContext *RunContext) HostHandlers {
	handlers := make(HostHandlers)
	appHosts := []AppHost{}
	err := runContext.Store.GetAll(&appHosts, 0, SentinelEnd)
	if len(appHosts) == 0 {
		return handlers
	}
	utils.HandleError(err)
	currentPlan, err := CurrentPlan(runContext)
	if err == ErrNoneFound {
		return handlers
	}
	utils.HandleError(err)

	for _, apphost := range appHosts {
		for _, appIDNeedRouting := range apphost.Apps {
			for _, sporeguid := range currentPlan.AppSchedule[appIDNeedRouting] {
				if runguids, ok := currentPlan.SporeSchedule[sporeguid.Sporeid]; ok {
					for runguid, runapp := range runguids {
						if runapp.ID != appIDNeedRouting {
							continue
						}
						spore, err := GetSpore(runContext, sporeguid.Sporeid)
						utils.HandleError(err)
						var host string
						if spore.ID == runContext.Config.MyMachineID {
							port := GetPortOn(runContext, spore, runapp, runguid)
							if port == "0" {
								utils.LogWarnF("Unable to loadbalance app %+v.", runapp)
							}
                                                        host = fmt.Sprintf("http://%v:%v", runContext.Config.DockerInterfaceIP, port)
                                                        fmt.Println(host);
						} else {
							// Todo(parham): check spores bound http port
							host = fmt.Sprintf("http://%v:80")
						}
						if _, ok := handlers[apphost.ID]; !ok {
							fwder, err := forward.New()
							utils.HandleError(err)
							rr, err := roundrobin.New(fwder)
							utils.HandleError(err)
							stream, _ := stream.New(rr, stream.Retry(`IsNetworkError() && Attempts() < 3`))
							handlers[apphost.ID] = streamRR{stream: stream, rr: rr}
						}
						serverUrl, err := url.Parse(host)
						utils.HandleError(err)
						handlers[apphost.ID].rr.UpsertServer(serverUrl)
					}
				}
			}

		}

	}
	return handlers
}
