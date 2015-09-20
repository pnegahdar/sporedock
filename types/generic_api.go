package types

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

type ApiEntity string

var ApiEntityHome ApiEntity = ""
var ApiEntityApp ApiEntity = "app"
var ApiEntityHost ApiEntity = "host"

var GenApiCreate = map[ApiEntity]Validable{ApiEntityApp: &App{}, ApiEntityHost: &AppHost{}}
var GenApiIndex = map[ApiEntity]Validable{ApiEntityApp: &App{}, ApiEntityHost: &AppHost{}}
var GenApiDelete = map[ApiEntity]Validable{ApiEntityApp: &App{}, ApiEntityHost: &AppHost{}}

func GenIndexAll(runContext *RunContext, genericTypeID ApiEntity) (valid interface{}, err error, statusCode int) {
	var validable interface{}
	switch genericTypeID {
	case ApiEntityApp:
		genericType := []App{}
		err := runContext.Store.GetAll(&genericType, 0, SentinelEnd)
		if err != nil {
			return nil, err, 400
		}
		validable = genericType
		return validable, nil, 200
	case ApiEntityHost:
		genericType := []AppHost{}
		err := runContext.Store.GetAll(&genericType, 0, SentinelEnd)
		if err != nil {
			return nil, err, 400
		}
		validable = genericType
		return validable, nil, 200
	default:
		return nil, ErrNotFound, 404
	}
}
