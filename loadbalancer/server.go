package loadbalancer

import (
	"github.com/mailgun/vulcan"
	"github.com/mailgun/vulcan/endpoint"
	"github.com/mailgun/vulcan/loadbalance/roundrobin"
	"github.com/mailgun/vulcan/location/httploc"
	"github.com/mailgun/vulcan/route"
	"github.com/pnegahdar/sporedock/utils"
	"net/http"
	"time"
)

func UpdateRouterFromManifest() {

}

func Serve() {

	// Create a round robin load balancer with some endpoints
	rr, err := roundrobin.NewRoundRobin()
	utils.HandleError(err)

	err = rr.AddEndpoint(endpoint.MustParseUrl("http://localhost:8000"))
	utils.HandleError(err)

	// Create a http location with the load balancer we've just added
	loc, err := httploc.NewLocation("loc1", rr)
	utils.HandleError(err)

	// Create a proxy server that routes all requests to "loc1"
	proxy, err := vulcan.NewProxy(&route.ConstRouter{Location: loc})
	utils.HandleError(err)

	// Proxy acts as http handler:
	server := &http.Server{
		Addr:           "localhost:8200",
		Handler:        proxy,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err = server.ListenAndServe()
	utils.HandleError(err)
	for {
		time.Sleep(time.Second * 5)
		UpdateRouterFromManifest()
	}
}
