package grunts

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"gopkg.in/tylerb/graceful.v1"
	"io/ioutil"
	"net/http"
	"time"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

type SporeAPI struct {
	runContext *types.RunContext
	stopCast   utils.SignalCast
}

func (sa SporeAPI) ProcName() string {
	return "SporeAPI"
}

func (sa SporeAPI) ShouldRun(runContext *types.RunContext) bool {
	return true
}

func (sa SporeAPI) Run(runContexnt *types.RunContext) {
	sa.runContext = runContexnt
	routes := Routes{
		Route{
			"Index",
			"GET",
			types.GetRoute(types.EntityTypeHome),
			sa.Home,
		},
		Route{
			"WebAppIndex",
			"GET",
			types.GetRoute(types.EntityTypeWebapp),
			sa.WebAppsIndex,
		},
		Route{
			"WebAppCreate",
			"POST",
			types.GetRoute(types.EntityTypeWebapp),
			sa.WebAppCreate,
		},
	}
	router := mux.NewRouter().StrictSlash(false)
	// Register routes
	for _, route := range routes {
		router.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(route.HandlerFunc)
	}
	srv := &graceful.Server{
		Timeout: 10 * time.Second,
		Server:  &http.Server{Addr: ":5000", Handler: router},
	}
	go func() {
		err := srv.ListenAndServe()
		utils.HandleError(err)
	}()
	<-sa.stopCast.Listen()
	srv.Stop(0)
}

func (sa SporeAPI) Stop() {
	sa.stopCast.Signal()
}

func jsonErrorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	json_string, marshall_err := utils.Marshall(types.Response{Error: err.Error(), StatusCode: 400})
	utils.HandleError(marshall_err)
	fmt.Fprint(w, json_string)

}

func bodyString(r *http.Request) string {
	body, _ := ioutil.ReadAll(r.Body)
	return string(body)
}

func datafromJsonRequest(body string) (string, error) {
	request := types.JsonRequest{}
	err := utils.Unmarshall(body, &request)
	if err != nil {
		return request.Data, types.ErrUnparsableRequest
	}
	return request.Data, nil

}

func jsonSuccessResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json_string, err := utils.Marshall(types.Response{Data: data, StatusCode: status})
	utils.HandleError(err)
	fmt.Fprint(w, json_string)

}

func (sa SporeAPI) Home(w http.ResponseWriter, r *http.Request) {
	data := "Welcome to Sporedock"
	jsonSuccessResponse(w, 200, data)
}

func (sa SporeAPI) WebAppsIndex(w http.ResponseWriter, r *http.Request) {
	webapps, err := cluster.GetAllWebApps(sa.runContext)
	if err != nil {
		jsonErrorResponse(w, err)
	} else {
		jsonSuccessResponse(w, 200, webapps)
	}
}

func (sa SporeAPI) WebAppCreate(w http.ResponseWriter, r *http.Request) {
	data, err := datafromJsonRequest(bodyString(r))
	if err != nil {
		jsonErrorResponse(w, err)
		return
	}
	webapp := &cluster.WebApp{}
	err = utils.Unmarshall(data, webapp)
	if err != nil {
		jsonErrorResponse(w, err)
		return
	}
	sa.runContext.Store.Set(webapp, webapp.ID, -1)
	if err != nil {
		jsonErrorResponse(w, err)
		return
	}
	jsonSuccessResponse(w, 200, webapp)
	return
}
