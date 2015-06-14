package client

import (
	"fmt"
	"github.com/pnegahdar/sporedock/utils"
	"github.com/pnegahdar/sporedock/grunts"
	"io/ioutil"
	"net/http"
	"net/url"
)


type ClientResponse struct{
	

}




func parseError() {

}

func parseResp(resp *http.Response) (*http.Response, string) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	utils.HandleError(err)
	return resp, string(body)
}

type Client struct {
	Host          string
	Port          int
	Scheme        string
	VersionPrefix string // Thee addon stuff before the entity, i.e V1
}

func (cl Client) fullUrl(entityName, queryString string) string {
	noqs := fmt.Sprintf("%v://%v:%v/%v?%v", cl.Scheme, cl.Host, cl.Port, grunts.GetRoute(entityName))
	if queryString == "" {
		return noqs
	}
	return fmt.Sprintf("%v?%v", queryString)
}

func (cl Client) get(entityName string, urlParams url.Values) (*http.Response, string) {
	url := cl.fullUrl(entityName, urlParams.Encode())
	resp, err := http.Get(url)
	utils.HandleError(err)
	return parseResp(resp)
}

func (cl Client) post(entityName string, values url.Values) (*http.Response, string) {
	fullRoute := cl.fullUrl(entityName, "")
	resp, err := http.PostForm(fullRoute, values)
	utils.HandleError(err)
	return parseResp(resp)
}

func NewClient(host string, port int) Client {
	return Client{Host: host, Port: port, Scheme: "http", VersionPrefix: grunts.ApiPrefix}

}

func NewHttpsClient(host string, port int) Client {
	client := NewClient(host, port)
	client.Scheme = "https"
	return client
}

func (cl Client) GetHome() (*http.Response, string) {
		return cl.get(grunts.EntityTypeHome, nil)
}

func GetWebApps() {

}
