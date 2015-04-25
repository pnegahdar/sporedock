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
)

type successResposne struct {
	Data interface{} `json:"data"`
	Code string      `json:"code"`
}

type errorRepsonse struct {
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

func (sa SporeAPI) Run(runContexnt *types.RunContext) {
	sa.runContext = runContexnt
	routes := Routes{
		Route{
			"Index",
			"GET",
			"/",
			sa.Home,
		},
		Route{
			"WebAppIndex",
			"GET",
			"/webapp/",
			sa.WebAppsIndex,
		},
		Route{
			"WebAppCreate",
			"POST",
			"/webapp/",
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

func jsonErrorResponse(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	error_body := map[string]string{"code": strconv.Itoa(status), "message": err.Error()}
	json_string, err := utils.Marshall(errorRepsonse{Error: error_body})
	utils.HandleError(err)
	fmt.Fprint(w, json_string)

}

func bodyString(r *http.Request) string {
	body, _ := ioutil.ReadAll(r.Body)
	return string(body)
}

func jsonSuccessResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json_string, err := utils.Marshall(successResposne{Data: data, Code: strconv.Itoa(status)})
	utils.HandleError(err)
	fmt.Fprint(w, json_string)

}

func (sa SporeAPI) Home(w http.ResponseWriter, r *http.Request) {
	data := "Welcome to Sporedock"
	jsonSuccessResponse(w, 200, data)
}

func (sa SporeAPI) WebAppsIndex(w http.ResponseWriter, r *http.Request) {
	webapps := cluster.GetAllWebApps(sa.runContext)
	jsonSuccessResponse(w, 200, webapps)
}

func (sa SporeAPI) WebAppCreate(w http.ResponseWriter, r *http.Request) {
	var webapp cluster.WebApp
	storable, err := webapp.FromString(bodyString(r))
	webapp = storable.(cluster.WebApp)
	fmt.Println(webapp)
	fmt.Println(err)
}
