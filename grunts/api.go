package grunts

import (
    "github.com/gorilla/mux"
    "github.com/pnegahdar/sporedock/utils"
    "net/http"
    "fmt"
    "github.com/pnegahdar/sporedock/cluster"
)


type Route struct {
    Name    string
    Method  string
    Pattern string
    HandlerFunc http.HandlerFunc
}

type Routes []Route

type SporeAPI struct {
    runContext *RunContext
}

func (sa SporeAPI) ProcName() string {
    return "SporeAPI"
}

func (sa SporeAPI) ShouldRun(runContext RunContext) bool {
    return true
}

func (sa SporeAPI) Run(runContexnt *RunContext) {
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
    for _, route := range (routes) {
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
    retType := cluster.WebApp{}
    storables := sa.runContext.store.GetAll(retType).([]cluster.WebApp)
    fmt.Println(storables)
}

