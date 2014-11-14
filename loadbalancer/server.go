package loadbalancer

import (
	"fmt"
	"github.com/mailgun/vulcan"
	"github.com/mailgun/vulcan/endpoint"
	"github.com/mailgun/vulcan/loadbalance/roundrobin"
	"github.com/mailgun/vulcan/location/httploc"
	"github.com/mailgun/vulcan/route"
	"github.com/mailgun/vulcan/route/hostroute"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/utils"
	"net/http"
	"strings"
	"time"
)

func UpdateRoutes(currentRoute *hostroute.HostRouter) {
	newHostRouter := hostroute.NewHostRouter()
	utils.LogDebug("Updating routes.")
	currentCluster := cluster.GetCurrentCluster()
	for _, webapp := range currentCluster.WebApps {
		rr, err := roundrobin.NewRoundRobin()
		resp, err := server.EtcdClient().Get(cluster.GetAppLocationKey(webapp.GetName()), true, false)
		if err != nil && strings.Index(err.Error(), "Key not found") != -1 {
			continue
		}
		for _, node := range resp.Node.Nodes {
			utils.LogDebug(fmt.Sprintf("Added host %v for app %v", node.Value, webapp.GetName()))
			err := rr.AddEndpoint(endpoint.MustParseUrl(node.Value))
			utils.HandleError(err)
		}
		loc, err := httploc.NewLocation(webapp.GetName(), rr)
		utils.HandleError(err)
		for _, hostname := range webapp.WebEndpoints {
			err = newHostRouter.SetRouter(hostname, &route.ConstRouter{Location: loc})
			utils.HandleError(err)
		}
		*currentRoute = *newHostRouter
	}
}

func WebServer(router *hostroute.HostRouter) {
	proxy, err := vulcan.NewProxy(router)
	utils.HandleError(err)
	server := &http.Server{
		Addr:           "localhost:8200",
		Handler:        proxy,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err = server.ListenAndServe()
	utils.HandleError(err)
}


func Run(){
	router := hostroute.NewHostRouter()
	go WebServer(router)
	for {
		time.Sleep(time.Second * 5)
		UpdateRoutes(router)
	}

}