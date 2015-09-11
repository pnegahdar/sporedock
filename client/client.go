package client

import (
	"errors"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type ClientResponse struct {
	HttpResponse *http.Response
	Content      string
	Response     types.Response
}

type Client struct {
	Host          string
	Port          int
	Scheme        string
	VersionPrefix string // Thee addon stuff before the entity, i.e V1
}

func (cl Client) fullGenUrl(route, queryString string) string {
	noqs := fmt.Sprintf("%v://%v:%v%v", cl.Scheme, cl.Host, cl.Port, route)
	if queryString == "" {
		return noqs
	}
	return fmt.Sprintf("%v?%v", noqs, queryString)
}

func (cl Client) get(url string, urlParams url.Values) (ClientResponse, *utils.BatchError) {
	batchError := &utils.BatchError{}
	resp, err := http.Get(url)
	batchError.Add(err)
	cr, err := parseResp(resp)
	batchError.Add(err)
	if cr.Response.Error != "" {
		batchError.Add(errors.New(cr.Response.Error))
	}
	return cr, batchError
}

func (cl Client) post(url string, jsonInterface interface{}) (ClientResponse, *utils.BatchError) {
	batchError := &utils.BatchError{}
	objstr, err := utils.Marshall(jsonInterface)
	batchError.Add(err)
	reqObject := types.JsonRequest{Data: objstr}
	body, err := utils.Marshall(reqObject)
	batchError.Add(err)
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	batchError.Add(err)
	cr, err := parseResp(resp)
	batchError.Add(err)
	if cr.Response.Error != "" {
		batchError.Add(errors.New(cr.Response.Error))
	}
	return cr, batchError
}

func (cl Client) delete(url string, urlParams url.Values) (ClientResponse, *utils.BatchError) {
	batchError := &utils.BatchError{}
	req := gorequest.New()
	resp, _, errs := req.Delete(url).End()
	for _, err := range errs {
		batchError.Add(err)
	}
	cr, err := parseResp(resp)
	batchError.Add(err)
	if cr.Response.Error != "" {
		batchError.Add(errors.New(cr.Response.Error))
	}
	return cr, batchError
}

func parseResp(resp *http.Response) (ClientResponse, error) {
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	cr := ClientResponse{HttpResponse: resp, Content: string(body), Response: types.Response{}}
	err := utils.Unmarshall(cr.Content, &cr.Response)
	return cr, err
}

func NewClient(host string, port int) *Client {
	return &Client{Host: host, Port: port, Scheme: "http", VersionPrefix: types.ApiPrefix}

}

func NewHttpsClient(host string, port int) *Client {
	client := NewClient(host, port)
	client.Scheme = "https"
	return client
}

func (cl Client) GetGeneric(unpack interface{}, genType string, id string) error {
	url := cl.fullGenUrl(types.GetGenApiRoute(genType, id), "")
	cr, batchErr := cl.get(url, nil)
	backToString, err := utils.Marshall(cr.Response.Data)
	batchErr.Add(err)
	err = utils.Unmarshall(backToString, unpack)
	batchErr.Add(err)
	return batchErr.Error()
}

func (cl Client) GetAllGeneric(unpack interface{}, genType string) error {
	url := cl.fullGenUrl(types.GetGenApiRoute(genType), "")
	cr, batchErr := cl.get(url, nil)
	backToString, err := utils.Marshall(cr.Response.Data)
	batchErr.Add(err)
	err = utils.Unmarshall(backToString, unpack)
	batchErr.Add(err)
	return batchErr.Error()
}

func (cl Client) DeleteGeneric(genType string, id string) error {
	url := cl.fullGenUrl(types.GetGenApiRoute(genType, id), "")
	_, batchErr := cl.delete(url, nil)
	return batchErr.Error()
}

func (cl Client) CreateGeneric(unpack interface{}, genType string) error {
	url := cl.fullGenUrl(types.GetGenApiRoute(genType), "")
	cr, batchErr := cl.post(url, unpack)
	backToString, err := utils.Marshall(cr.Response.Data)
	batchErr.Add(err)
	err = utils.Unmarshall(backToString, unpack)
	batchErr.Add(err)
	return batchErr.Error()
}

//WEBAPP
func (cl Client) GetApps() ([]types.App, error) {
	webapps := []types.App{}
	err := cl.GetAllGeneric(&webapps, types.EntityTypeApp)
	return webapps, err
}

func (cl Client) GetApp(id string) (*types.App, error) {
	webapps := &types.App{}
	err := cl.GetGeneric(webapps, types.EntityTypeApp, id)
	return webapps, err
}

func (cl Client) CreateApp(webapp *types.App) error {
	err := cl.CreateGeneric(webapp, types.EntityTypeApp)
	return err
}

func (cl Client) DeleteApp(id string) error {
	err := cl.DeleteGeneric(types.EntityTypeApp, id)
	return err
}
