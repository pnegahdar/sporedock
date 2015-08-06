package client

import (
	"errors"
	"fmt"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type ClientResponse struct {
	HttpReposne       *http.Response
	Content           string
	SporeDockResponse types.Response
}

func parseResp(resp *http.Response) ClientResponse {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	utils.HandleError(err)
	cr := ClientResponse{HttpReposne: resp, Content: string(body), SporeDockResponse: types.Response{}}
	err = utils.Unmarshall(cr.Content, &cr.SporeDockResponse)
	utils.HandleError(err)
	return cr
}

type Client struct {
	Host          string
	Port          int
	Scheme        string
	VersionPrefix string // Thee addon stuff before the entity, i.e V1
}

func (cl Client) fullUrl(entityName, queryString string) string {
	noqs := fmt.Sprintf("%v://%v:%v%v", cl.Scheme, cl.Host, cl.Port, types.GetGenApiRoute(entityName))
	if queryString == "" {
		return noqs
	}
	return fmt.Sprintf("%v?%v", queryString)
}

func (cl Client) get(entityName string, urlParams url.Values) ClientResponse {
	url := cl.fullUrl(entityName, urlParams.Encode())
	resp, err := http.Get(url)
	utils.HandleError(err)
	return parseResp(resp)
}

func (cl Client) postjson(obj interface{}) ClientResponse {
	fullRoute := cl.fullUrl("webapp", "")
	objstr, err := utils.Marshall(obj)
	utils.HandleError(err)
	reqObject := types.JsonRequest{Data: objstr}
	body, err := utils.Marshall(reqObject)
	utils.HandleError(err)
	resp, err := http.Post(fullRoute, "application/json", strings.NewReader(body))
	utils.HandleError(err)
	return parseResp(resp)
}

func NewClient(host string, port int) *Client {
	return &Client{Host: host, Port: port, Scheme: "http", VersionPrefix: types.ApiPrefix}

}

func NewHttpsClient(host string, port int) *Client {
	client := NewClient(host, port)
	client.Scheme = "https"
	return client
}

func (cl Client) GetHome() (string, error) {
	resp := cl.get(types.EntityTypeHome, nil)
	if resp.SporeDockResponse.IsError() {
		return "", errors.New(resp.SporeDockResponse.Error)
	}
	return resp.Content, nil
}

func (cl Client) GetWebApps() ([]cluster.WebApp, error) {
	webapps := []cluster.WebApp{}
	resp := cl.get(types.EntityTypeWebapp, nil)
	if resp.SporeDockResponse.IsError() {
		return webapps, errors.New(resp.SporeDockResponse.Error)

	}
	toConvert := resp.SporeDockResponse.Data.([]interface{})
	for _, wa := range toConvert {
		webapps = append(webapps, wa.(cluster.WebApp))
	}
	return webapps, nil
}

func (cl Client) CreateWebApp(webapp cluster.WebApp) (cluster.WebApp, error) {
	resp := cl.postjson(webapp)
	if resp.SporeDockResponse.IsError() {
		return webapp, errors.New(resp.SporeDockResponse.Error)
	}
	return webapp, nil
}
