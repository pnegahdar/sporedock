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

func frontendDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Join(path.Dir(filename), "../frontend")
}

func frontendSubDir(addon...string) string {
	parts := []string{frontendDir()}
	for _, v := range (addon) {
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

var TypeMap = map[string]types.Validable{"webapp" : cluster.WebApp{}}

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

func (sa SporeAPI) GenericTypeGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	genericTypeID := vars["type"]
	objectID := vars["id"]
	creatable, ok := TypeMap[genericTypeID]
	if !ok {
		jsonErrorResponse(w, types.ErrNotFound, 404)
		return
	}
	err := sa.runContext.Store.Get(&creatable, objectID)
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	jsonSuccessResponse(w, 200, creatable)
	return
}

func (sa SporeAPI) GenericTypeCreate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	genericTypeID := vars["type"]
	data, err := datafromJsonRequest(bodyString(r))
	var validable types.Validable
	switch genericTypeID{
	case "webapp":
		toSave := cluster.WebApp{}
		err = utils.Unmarshall(data, &toSave)
		if err != nil {
			jsonErrorResponse(w, err, 400)
			return
		}
		validable = toSave
	default:
		jsonErrorResponse(w, types.ErrNotFound, 404)
		return
	}
	err = validable.Validate(sa.runContext)
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	err = sa.runContext.Store.Set(&validable, validable.GetID(), -1)
	if err != nil {
		jsonErrorResponse(w, err, 400)
		return
	}
	jsonSuccessResponse(w, 200, validable)
	return
}

func (sa SporeAPI) GenericTypeDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	genericTypeID := vars["type"]
	objectID := vars["id"]
	creatable, ok := TypeMap[genericTypeID]
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
	return
}
