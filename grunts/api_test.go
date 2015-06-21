package grunts

import (
// "fmt"
	"github.com/pnegahdar/sporedock/client"
// "github.com/pnegahdar/sporedock/utils"
	"github.com/stretchr/testify/suite"
	"testing"
	"fmt"
	"github.com/pnegahdar/sporedock/cluster"
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
	go CreateAndRun()
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

func (suite *ApiTestSuite) TestCreateWebapp(){
	toCreate := cluster.NewWebApp("TESTWEBAPP", TestImage, 8000)
	webapp, err := suite.Client.CreateWebApp(*toCreate)
	suite.Nil(err)
	suite.Equal(webapp, *toCreate)
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}


// func TestCreateWebapp(t *testing.T) {
// 	// Test No ID
// 	resp, _ := testPost(t, EntityTypeWebapp, url.Values{})
//
// 	assert.Equal(t, resp.StatusCode, 400)
//
// }
