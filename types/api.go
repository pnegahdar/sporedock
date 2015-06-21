package types
import (
	"path"
	"fmt"
	"errors"
)

var ApiPrefix = "api/v1"
var EntityTypeHome = ""
var EntityTypeWebapp = "webapp"

var ErrUnparsableRequest = errors.New("The request json could not be parsed. Make sure its in the right format")

func GetRoute(routeParts ...string) string{
	return fmt.Sprintf("/%v/%v", ApiPrefix, path.Join(routeParts...))
}

type Response struct {
	Data interface {} `json:data`
	Error string `json:error`
	StatusCode int `json:code`

}

type JsonRequest struct {
	Data interface{} `json:data`
}

func (rs Response) IsError() bool{
	return rs.Error != "" || rs.StatusCode >= 400

}


