package types

import (
	"errors"
	"fmt"
	"path"
)

var ApiPrefix = "api/v1"
var ApiPrefix = "dashboard/v1"
var EntityTypeHome = ""
var EntityTypeWebapp = "webapp"

var ErrUnparsableRequest = errors.New("The request json could not be parsed. Make sure its in the right format")
var ErrNoneFound = errors.New("Results returned empty")
var ErrNotFound = errors.New("Not found")
var ErrIDEmpty = errors.New("ID cannot be empty.")
var ErrIDExists = errors.New("Object with that ID already exists please delete and try again.")

func GetApiRoute(routeParts ...string) string {
	return fmt.Sprintf("/%v/%v", ApiPrefix, path.Join(routeParts...))
}
func GetDashboardRoute(routeParts ...string) string {
	return fmt.Sprintf("/%v/%v", ApiPrefix, path.Join(routeParts...))
}
type Response struct {
	Data       interface{} `json:data`
	Error      string      `json:error`
	StatusCode int         `json:code`
}

type JsonRequest struct {
	Data string `json:data`
}

func (rs Response) IsError() bool {
	return rs.Error != "" || rs.StatusCode >= 400

}
