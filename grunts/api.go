package grunts

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pnegahdar/sporedock/utils"
    "github.com/pnegahdar/sporedock/cluster"
	"net/http"
    "github.com/pnegahdar/sporedock/types"
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
			"/",
			sa.Home,
		},
		Route{
			"WebAppIndex",
			"GET",
			"/webapp/",
			sa.WebAppsIndex,
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

func jsonSuccessResponse(w http.ResponseWriter, jsonString string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, jsonString)

}

func (sa SporeAPI) Home(w http.ResponseWriter, r *http.Request) {
	jsonSuccessResponse(w, "{\"data\" : \"Welcome to SporeDock\" }")
}

func (sa SporeAPI) WebAppsIndex(w http.ResponseWriter, r *http.Request) {
    webapps := cluster.GetAllWebApps(sa.runContext)
    fmt.Println(webapps)
    // fmt.Println(sa.runContext.Store.GetAll(webApp))

}
