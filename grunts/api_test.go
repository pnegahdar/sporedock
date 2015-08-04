package grunts


import (
// "fmt"
	"github.com/pnegahdar/sporedock/client"
// "github.com/pnegahdar/sporedock/utils"
	"fmt"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/stretchr/testify/suite"
	"testing"
)

var TestImage = "ubuntu"

func handleTestError(t *testing.T, err error) {
	if err != nil {
		t.Error(err.Error())
	}
}

type ApiTestSuite struct {
	suite.Suite
	Client *client.Client
}

func (suite *ApiTestSuite) SetupSuite() {
	suite.Client = client.NewClient("localhost", 5000)
}

func (suite *ApiTestSuite) TestHome() {
	resp, err := suite.Client.GetHome()
	suite.Nil(err)
	suite.Contains(resp, "Welcome", "Welcome not found in home body.")
}

func (suite *ApiTestSuite) TestNoWebapps() {
	webapps, err := suite.Client.GetWebApps()
	suite.Nil(err)
	fmt.Println(webapps)
	suite.Len(webapps, 0)
}

func (suite *ApiTestSuite) TestCreateWebapp() {
	toCreate := cluster.NewWebApp("TESTWEBAPP", TestImage, 8000)
	webapp, err := suite.Client.CreateWebApp(*toCreate)
	suite.Nil(err)
	suite.Equal(webapp, *toCreate)
}

func TestApiTestSuite(t *testing.T) {
	registry := CreateAndRun("redis://localhost:6379", "testGroup", "myMachine", "127.0.0.1")
	suite.Run(t, new(ApiTestSuite))
	registry.Stop()
}

// func TestCreateWebapp(t *testing.T) {
// 	// Test No ID
// 	resp, _ := testPost(t, EntityTypeWebapp, url.Values{})
//
// 	assert.Equal(t, resp.StatusCode, 400)
//
// }
