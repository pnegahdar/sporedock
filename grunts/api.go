package grunts

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
	"net/http"
	"path"
	"runtime"
	"runtime/debug"
	"sync"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var genCreate = map[string]types.Validable{"webapp": &cluster.WebApp{}}
var genIndex = map[string]types.Validable{"webapp": &cluster.WebApp{}}
var genDelete = map[string]types.Validable{"webapp": &cluster.WebApp{}}

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

func (sa *SporeAPI) Run(runContext *types.RunContext) {
	sa.Lock()
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
			types.GetGenApiRoute("{type}"),
			sa.GenericTypeIndex,
		},
		Route{
			"GenericTypeCreate",
			"POST",
			types.GetGenApiRoute("{type}"),
			sa.GenericTypeCreate,
		},
		Route{
			"GenericTypeGet",
			"GET",
			types.GetGenApiRoute("{type}", "{id}"),
			sa.GenericTypeGet,
		},
		Route{
			"GenericTypeDelete",
			"DELETE",
			types.GetGenApiRoute("{type}", "{id}"),
			sa.GenericTypeDelete,
		},
	}
	// Register API routes
	for _, route := range routes {
		sa.runContext.WebServerRouter.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(route.HandlerFunc)
	}
	// Dash Routes
	staticRoute := types.GetDashboardRoute("static")
	staticHandler := http.StripPrefix(staticRoute, http.FileServer(http.Dir(frontendSubDir("static"))))
	sa.runContext.WebServerRouter.Methods("GET").PathPrefix(staticRoute).Name("DashboardStaticFiles").Handler(staticHandler)
	sa.runContext.WebServerRouter.Methods("GET").PathPrefix(types.GetDashboardRoute()).Name("DashboardIndex").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, frontendSubDir("index.html"))
		})
	exit, _ := sa.stopCast.Listen()
	sa.Unlock()
	<-exit
}

// Todo: make sure stop works without pointer receivers?
func (sa *SporeAPI) Stop() {
	sa.stopCastMu.Lock()
	defer sa.stopCastMu.Unlock()
	sa.stopCast.Signal()
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
	var validable interface{}
	switch genericTypeID {
	case "webapp":
		genericType := []cluster.WebApp{}
		err := sa.runContext.Store.GetAll(&genericType, 0, types.SentinelEnd)
		if err != nil {
			jsonErrorResponse(w, err, 400)
			return
		}
		validable = genericType
	default:
		jsonErrorResponse(w, types.ErrNotFound, 404)
		return
	}
	jsonSuccessResponse(w, 200, validable)
}

func (sa SporeAPI) GenericTypeGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	genericTypeID := vars["type"]
	objectID := vars["id"]
	indexable, ok := genIndex[genericTypeID]
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
	genericTypeID := vars["type"]
	data, err := datafromJsonRequest(bodyString(r))
	creatable, ok := genCreate[genericTypeID]
	if !ok {
		jsonErrorResponse(w, types.ErrNotFound, 404)
		return
	}
	err = utils.Unmarshall(data, &creatable)
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
	genericTypeID := vars["type"]
	objectID := vars["id"]
	creatable, ok := genDelete[genericTypeID]
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
