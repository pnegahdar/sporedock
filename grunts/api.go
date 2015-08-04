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
	"runtime"
	"path"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func frontendDir() string{
	_, filename, _, _ := runtime.Caller(1)
	return path.Join(path.Dir(filename), "../frontend")
}

func frontendSubDir(addon...string) string{
	parts := []string{frontendDir()}
	for _, v := range(addon){
		parts = append(parts, v)
	}
	return path.Join(parts...)
}

type SporeAPI struct {
	runContext *types.RunContext
	stopCast   utils.SignalCast
}

func (sa SporeAPI) ProcName() string {
	return "SporeAPI"
}

var typeMap = map[string]types.Creatable{"webapp" : cluster.WebApp{}}

func (sa SporeAPI) ShouldRun(runContext *types.RunContext) bool {
	return true
}

func (sa SporeAPI) Run(runContext *types.RunContext) {
	sa.runContext = runContext
	routes := Routes{
		// API
		Route{
			"Index",
			"GET",
			types.GetApiRoute(types.EntityTypeHome),
			sa.Home,
		},
		Route{
			"GenericTypeIndex",
			"GET",
			types.GetApiRoute("gen", "{type}"),
			sa.GenericTypeIndex,
		},
		Route{
			"GenericTypeCreate",
			"POST",
			types.GetApiRoute("gen", "{type}"),
			sa.GenericTypeCreate,
		},
	}
	router := mux.NewRouter()
	// Register API routes
	for _, route := range routes {
		router.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(route.HandlerFunc)
	}
	// Dash Routes
	staticRoute := types.GetDashboardRoute("static")
	staticHandler := http.StripPrefix(staticRoute, http.FileServer(http.Dir(frontendSubDir("static"))))
	router.Methods("GET").PathPrefix(staticRoute).Name("DashboardStaticFiles").Handler(staticHandler)
	router.Methods("GET").PathPrefix(types.GetDashboardRoute()).Name("DashboardIndex").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, frontendSubDir("index.html"))
		})
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
	switch genericTypeID {
	case "webapp":
		genericType := []cluster.WebApp{}
		err := sa.runContext.Store.GetAll(&genericType, 0, types.SentinelEnd)
		if err != nil {
			jsonErrorResponse(w, err, 400)
		} else {
			jsonSuccessResponse(w, 200, genericType)
		}
		return
	default:
		jsonErrorResponse(w, types.ErrNotFound, 404)
		return
	}
}

func (sa SporeAPI) GenericTypeCreate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	genericTypeID := vars["type"]
	data, err := datafromJsonRequest(bodyString(r))
	creatable, ok := typeMap[genericTypeID]
	if !ok {
		jsonErrorResponse(w, types.ErrNotFound, 404)
		return
	}
	created, err := creatable.Create(sa.runContext, data)
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	jsonSuccessResponse(w, 200, created)
	return
}
