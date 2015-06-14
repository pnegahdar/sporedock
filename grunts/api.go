package grunts

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
	"net/http"
	"strconv"
	"path"
)

var ApiPrefix = "api/v1"
var EntityTypeHome = ""
var EntityTypeWebapp = "webapp"

type successResponse struct {
	Data interface{} `json:"data"`
	Code string      `json:"code"`
}

type errorResponse struct {
	Error map[string]string `json:"error"`
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

type SporeAPI struct {
	runContext *types.RunContext
}

func (sa SporeAPI) ProcName() string {
	return "SporeAPI"
}

func (sa SporeAPI) ShouldRun(runContext types.RunContext) bool {
	return true
}

func GetRoute(routeParts ...string) string{
	return fmt.Sprintf("/%v/%v", ApiPrefix, path.Join(routeParts...))
}

func (sa SporeAPI) Run(runContexnt *types.RunContext) {
	sa.runContext = runContexnt
	routes := Routes{
		Route{
			"Index",
			"GET",
			GetRoute(EntityTypeHome),
			sa.Home,
		},
		Route{
			"WebAppIndex",
			"GET",
			GetRoute(EntityTypeWebapp),
			sa.WebAppsIndex,
		},
		Route{
			"WebAppCreate",
			"POST",
			GetRoute(EntityTypeWebapp),
			sa.WebAppCreate,
		},
	}
	router := mux.NewRouter().StrictSlash(false)
	// Register routes
	for _, route := range routes {
		router.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(route.HandlerFunc)
	}
	err := http.ListenAndServe(":5000", router)
	utils.HandleError(err)
}

func jsonErrorResponse(w http.ResponseWriter, err interface{}) {
	err_wrapped := types.RewrapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err_wrapped.Status)
	error_body := map[string]string{"code": strconv.Itoa(err_wrapped.Status), "message": err_wrapped.Error.Error()}
	json_string, marshall_err := utils.Marshall(errorResponse{Error: error_body})
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
	json_string, err := utils.Marshall(successResponse{Data: data, Code: strconv.Itoa(status)})
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
	var webapp cluster.WebApp
	storable, err := webapp.FromString(bodyString(r), sa.runContext)
	webapp = storable.(cluster.WebApp)
	fmt.Println(webapp)
	if err != nil {
		jsonErrorResponse(w, err)
	}
}
