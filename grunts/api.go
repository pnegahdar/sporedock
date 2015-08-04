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

var typeMap = map[string]interface{}{
	types.EntityTypeWebapp : cluster.WebApp{}}

var typeMapSlice = map[string]interface{}{
	types.EntityTypeWebapp : []cluster.WebApp{}}

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
			"GenericTypeIndex",
			"GET",
			types.GetRoute("gen", "{type}"),
			sa.GenericTypeIndex,
		},
		Route{
			"WebAppCreate",
			"POST",
			types.GetRoute("gen", "{type}", "{id}"),
			sa.GenericTypeCreate,
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

func jsonErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json_string, marshall_err := utils.Marshall(types.Response{Error: err.Error(), StatusCode: statusCode})
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

func (sa SporeAPI) GenericTypeIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	genericTypeID := vars["type"]
	genericType, ok := typeMapSlice[genericTypeID]
	if !ok {
		jsonErrorResponse(w, types.ErrNotFound, 404)
		return
	}
	err := sa.runContext.Store.GetAll(genericType, 0, types.SentinelEnd)
	if err != nil {
		jsonErrorResponse(w, err, 400)
	} else {
		jsonSuccessResponse(w, 200, genericType)
	}
}

func (sa SporeAPI) GenericTypeCreate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	genericTypeID := vars["type"]
	typeID := vars["id"]
	genericType, ok := typeMap[genericTypeID]
	if !ok {
		jsonErrorResponse(w, types.ErrNotFound, 404)
		return
	}
	data, err := datafromJsonRequest(bodyString(r))
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	err = utils.Unmarshall(data, genericType)
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	err = nil // Todo(parham): call validate method here.
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	sa.runContext.Store.Set(genericType, typeID, -1)
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	jsonSuccessResponse(w, 200, genericType)
	return
}
