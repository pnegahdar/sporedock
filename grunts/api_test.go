package grunts

import (
	"fmt"
	"github.com/pnegahdar/sporedock/client"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/url"
	"os"
	"testing"
)

func handleTestError(t *testing.T, err error) {
	if err != nil {
		t.Error(err.Error())
	}
}

func createTestClient() client.Client {
	return  client.NewClient("localhost", 80)
}

func TestMain(m *testing.M) {
	go CreateAndRun()
	os.Exit(m.Run())
}

type 

func TestHome(t *testing.T) {
	cr =
	resp, content := testGet(t, EntityTypeHome, url.Values{})
	assert.Equal(t, resp.StatusCode, 200, "Status code was not 200")
	assert.Contains(t, content, "Welcome", "Welcome not found in home body.")
}

func TestNoWebapps(t *testing.T) {
	resp, content := testGet(t, EntityTypeWebapp, url.Values{})
	assert.Equal(t, resp.StatusCode, 200, "Status code was not 200")
	sresp := successResponse{}
	t.Log(content)
	err := utils.Unmarshall(content, &sresp)
	handleTestError(t, err)
	webapps := []cluster.WebApp{}
	toConvert := sresp.Data.([]interface{})
	for _, wa := range toConvert {
		webapps = append(webapps, wa.(cluster.WebApp))
	}

	fmt.Println(webapps)
	assert.Equal(t, webapps, []cluster.WebApp{})
}

func TestCreateWebapp(t *testing.T) {
	// Test No ID
	resp, _ := testPost(t, EntityTypeWebapp, url.Values{})

	assert.Equal(t, resp.StatusCode, 400)

}
