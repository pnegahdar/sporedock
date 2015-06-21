package grunts

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
	"net/http"
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
	err := http.ListenAndServe(":5000", router)
	utils.HandleError(err)
}

func jsonErrorResponse(w http.ResponseWriter, err interface{}) {
	err_wrapped := types.RewrapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err_wrapped.Status)
	json_string, marshall_err := utils.Marshall(types.Response{Error: err_wrapped.Error.Error(), StatusCode: err_wrapped.Status})
	utils.HandleError(marshall_err)
	fmt.Fprint(w, json_string)

}

func bodyString(r *http.Request) string {
	body, _ := ioutil.ReadAll(r.Body)
	return string(body)
}

func parseJsonRequest(body string) ([]interface{}, error){
	request := types.JsonRequest{}
	err := utils.Unmarshall(body, &request)
	if err != nil{
		return request.Data, types.ErrUnparsableRequest
	}
	reqItems, ok := request.Data.([]interface{})
	if !ok{
		return request.Data,
	}

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
	var webapp cluster.WebApp
	storable, err := webapp.FromString(bodyString(r), sa.runContext)
	if err != nil {
		jsonErrorResponse(w, err)
	}
	webapp = storable.(cluster.WebApp)
	fmt.Println(webapp)
}
