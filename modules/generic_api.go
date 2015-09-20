package modules

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
	"net/http"
	"path"
	"runtime"
	"runtime/debug"
	"sync"
)

func frontendDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Join(path.Dir(filename), "../frontend")
}

func frontendSubDir(addon ...string) string {
	parts := []string{frontendDir()}
	for _, v := range addon {
		parts = append(parts, v)
	}
	return path.Join(parts...)
}

type SporeAPI struct {
	sync.Mutex
	initOnce   sync.Once
	runContext *types.RunContext
	stopCast   utils.SignalCast
	stopCastMu sync.Mutex
}

func (sa SporeAPI) ProcName() string {
	return "SporeAPI"
}

func (sa SporeAPI) ShouldRun(runContext *types.RunContext) bool {
	return true
}

func (sa *SporeAPI) Init(runContext *types.RunContext) {
	sa.initOnce.Do(func() {
		sa.runContext = runContext
		sa.setupRoutes()
	})
}

func (sa *SporeAPI) Run(runContext *types.RunContext) {
	exit, _ := sa.stopCast.Listen()
	<-exit
}

// Todo: make sure stop works without pointer receivers?
func (sa *SporeAPI) Stop() {
	sa.stopCastMu.Lock()
	defer sa.stopCastMu.Unlock()
	sa.stopCast.Signal()
}

func (sa *SporeAPI) setupRoutes() {
	routes := types.Routes{
		// API
		types.Route{
			"Index",
			"GET",
			types.GetApiRoute(string(types.ApiEntityHome)),
			sa.Home,
		},
		types.Route{
			"GenericTypeIndex",
			"GET",
			types.GetGenApiRoute("{type}"),
			sa.GenericTypeIndex,
		},
		types.Route{
			"GenericTypeCreate",
			"POST",
			types.GetGenApiRoute("{type}"),
			sa.GenericTypeCreate,
		},
		types.Route{
			"GenericTypeGet",
			"GET",
			types.GetGenApiRoute("{type}", "{id}"),
			sa.GenericTypeGet,
		},
		types.Route{
			"GenericTypeDelete",
			"DELETE",
			types.GetGenApiRoute("{type}", "{id}"),
			sa.GenericTypeDelete,
		},
	}
	// Register API routes
	for _, route := range routes {
		sa.runContext.WebServerManager.WebServerRouter.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(route.HandlerFunc)
	}
	// Dash Routes
	staticRoute := types.GetDashboardRoute("static")
	staticHandler := http.StripPrefix(staticRoute, http.FileServer(http.Dir(frontendSubDir("static"))))
	sa.runContext.WebServerManager.WebServerRouter.Methods("GET").PathPrefix(staticRoute).Name("DashboardStaticFiles").Handler(staticHandler)
	sa.runContext.WebServerManager.WebServerRouter.Methods("GET").PathPrefix(types.GetDashboardRoute()).Name("DashboardIndex").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, frontendSubDir("index.html"))
		})
}
func jsonErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	tb := string(debug.Stack())
	json_string, marshall_err := utils.Marshall(types.Response{Error: err.Error(), StatusCode: statusCode, ErrorTB: tb})
	utils.HandleError(marshall_err)
	fmt.Fprint(w, json_string)

}

func bodyString(r *http.Request) string {
	body, _ := ioutil.ReadAll(r.Body)
	return string(body)
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
	genericTypeID := types.ApiEntity(vars["type"])
	validable, err, statusCode := types.GenIndexAll(sa.runContext, genericTypeID)
	if err != nil {
		jsonErrorResponse(w, err, statusCode)
		return
	}
	jsonSuccessResponse(w, 200, validable)
	return
}

func (sa SporeAPI) GenericTypeGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	genericTypeID := types.ApiEntity(vars["type"])
	objectID := vars["id"]
	indexable, ok := types.GenApiIndex[genericTypeID]
	if !ok {
		jsonErrorResponse(w, types.ErrNotFound, 404)
		return
	}
	err := sa.runContext.Store.Get(&indexable, objectID)
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	jsonSuccessResponse(w, 200, indexable)
}

func (sa SporeAPI) GenericTypeCreate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	genericTypeID := types.ApiEntity(vars["type"])
	creatable, ok := types.GenApiCreate[genericTypeID]
	if !ok {
		jsonErrorResponse(w, types.ErrNotFound, 404)
		return
	}
	err := utils.Unmarshall(bodyString(r), &creatable)
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	err = creatable.Validate(sa.runContext)
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	err = sa.runContext.Store.Set(&creatable, creatable.GetID(), -1)
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	jsonSuccessResponse(w, 200, creatable)
}

func (sa SporeAPI) GenericTypeDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	genericTypeID := types.ApiEntity(vars["type"])
	objectID := vars["id"]
	creatable, ok := types.GenApiDelete[genericTypeID]
	if !ok {
		jsonErrorResponse(w, types.ErrNotFound, 404)
		return
	}
	err := sa.runContext.Store.Delete(creatable, objectID)
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	jsonSuccessResponse(w, 200, true)
}
