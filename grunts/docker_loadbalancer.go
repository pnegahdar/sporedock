package grunts

//
//import (
//	"fmt"
//	"github.com/mailgun/vulcan"
//	"github.com/mailgun/vulcan/endpoint"
//	"github.com/mailgun/vulcan/loadbalance/roundrobin"
//	"github.com/mailgun/vulcan/location/httploc"
//	"github.com/mailgun/vulcan/route"
//	"github.com/mailgun/vulcan/route/hostroute"
//	"github.com/pnegahdar/sporedock/cluster"
//	"github.com/pnegahdar/sporedock/server"
//	"github.com/pnegahdar/sporedock/utils"
//	"net/http"
//	"strings"
//	"time"
//)
//
//func UpdateRoutes(currentRoute *hostroute.HostRouter) {
//	newHostRouter := hostroute.NewHostRouter()
//	utils.LogDebug("Updating routes.")
//	currentCluster := cluster.GetCurrentCluster()
//	for _, webapp := range currentCluster.WebApps {
//		rr, err := roundrobin.NewRoundRobin()
//		resp, err := server.EtcdClient().Get(cluster.GetAppLocationKey(webapp.GetName()), true, false)
//		if err != nil && strings.Index(err.Error(), "Key not found") != -1 {
//			continue
//		}
//		for _, node := range resp.Node.Nodes {
//			utils.LogDebug(fmt.Sprintf("Added host %v for app %v", node.Value, webapp.GetName()))
//			err := rr.AddEndpoint(endpoint.MustParseUrl(node.Value))
//			utils.HandleError(err)
//		}
//		loc, err := httploc.NewLocation(webapp.GetName(), rr)
//		utils.HandleError(err)
//		for _, hostname := range webapp.WebEndpoints {
//			err = newHostRouter.SetRouter(hostname, &route.ConstRouter{Location: loc})
//			utils.HandleError(err)
//		}
//		*currentRoute = *newHostRouter
//	}
//}
//
//func WebServer(router *hostroute.HostRouter) {
//	proxy, err := vulcan.NewProxy(router)
//	utils.HandleError(err)
//	server := &http.Server{
//		Addr:           "localhost:8009",
//		Handler:        proxy,
//		ReadTimeout:    10 * time.Second,
//		WriteTimeout:   10 * time.Second,
//		MaxHeaderBytes: 1 << 20,
//	}
//	err = server.ListenAndServe()
//	utils.HandleError(err)
//}
//
//func Run() {
//	router := hostroute.NewHostRouter()
//	go WebServer(router)
//	for {
//		time.Sleep(time.Second * 5)
//		UpdateRoutes(router)
//	}
//
//}
//func CleanupLocations() {
//	currentCluster := cluster.GetCurrentCluster()
//	currentManifest := cluster.GetCurrentManifest()
//	spores := []string{}
//	for _, spore := range currentManifest {
//		spores = append(spores, spore.Spore.Name)
//	}
//	// Remove APPS DNE
//	appNames := []string{}
//	for _, app := range currentCluster.IterApps() {
//		appNames = append(appNames, app.GetName())
//	}
//	store := store.GetStore()
//	resp, err := store.GetKey(AppLocationsKey)
//	utils.HandleError(err)
//	if resp == "" {
//		return
//	}
//	utils.HandleError(err)
//	for _, node := range resp.Noe.Nodes {
//		appName := pathLastPart(node.Key)
//		if !In(appNames, appName) {
//			utils.LogDebug(fmt.Sprintf("App %v no longer exists removing loc.", appName))
//			_, err := server.EtcdClient().Delete(node.Key, true)
//			utils.HandleError(err)
//		}
//	}
//	// Remove Machines DNE
//	for _, app := range currentCluster.IterApps() {
//		keyName := cluster.GetAppLocationKey(app.GetName())
//		resp, err := server.EtcdClient().Get(keyName, true, false)
//		if err != nil && strings.Index(err.Error(), "Key not found") != -1 {
//			continue
//		}
//		utils.HandleError(err)
//		for _, node := range resp.Node.Nodes {
//			machineName := pathLastPart(node.Key)
//			if !In(spores, machineName) {
//				utils.LogDebug(fmt.Sprintf("Machine %v no longer exists removing loc.", machineName))
//				_, err := server.EtcdClient().Delete(node.Key, true)
//				utils.HandleError(err)
//			}
//		}
//	}
//
//}
//
//func UpdateLocations(appNames []string) {
//	dc := CachedDockerClient()
//	store := store.GetStore()
//	mySpore := store.GetMe()
//	locations := cluster.GetCurrentLBSet()
//	for _, appName := range appNames {
//		resp, err := dc.InspectContainer(appName)
//		utils.HandleError(err)
//		// Remove dead app
//		if !resp.State.Running {
//			_, err := server.EtcdClient().Delete(keyName, true)
//			if err != nil && strings.Index(err.Error(), "Key not found") != -1 {
//				continue
//			}
//			utils.LogDebug(fmt.Sprintf("Removed dead app location %v", appName))
//			utils.HandleError(err)
//			continue
//		}
//		bindings := resp.NetworkSettings.Ports
//		for k, v := range bindings {
//			if k == "80/tcp" {
//				//Todo(parham): Only allows for one per node
//				location := mySpore.GetPortLocation(v[0].HostPort)
//				_, err := server.EtcdClient().Set(keyName, location, 0)
//				utils.HandleError(err)
//			}
//		}
//	}
//}

import (
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"sync"
	"time"
	"fmt"
)

var updateEndpointsEvery = time.Millisecond * 1000

type LoadBalancer struct {
	initOnce   sync.Once
	stopCast   utils.SignalCast
	runContext *types.RunContext
}

func (lb *LoadBalancer) init(runContext *types.RunContext) {
	lb.initOnce.Do(func() {
		lb.runContext = runContext
	})
}

func (lb *LoadBalancer) ProcName() string {
	return "LoadBalancer"
}

func (lb *LoadBalancer) Stop() {
	lb.stopCast.Signal()
}

func (lb *LoadBalancer) Run(runContext *types.RunContext) {
	exit, _ := lb.stopCast.Listen()
	lb.init(runContext)
	for {
		select {
		case <-time.After(time.Second * 10):
			fmt.Println("YO")
		case <-exit:
			return
		}
	}
}

func (lb *LoadBalancer) ShouldRun(runContext *types.RunContext) bool {
	return true
}

