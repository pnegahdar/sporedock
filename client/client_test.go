package client

import (
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/modules"
	"github.com/pnegahdar/sporedock/types"
	"github.com/stretchr/testify/suite"
	"sort"
	"strconv"
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
	runContext *types.RunContext
	Client     *Client
}

func (suite *ApiTestSuite) cleanup() {
	err := suite.runContext.Store.DeleteAll(types.App{})
	suite.Nil(err)
}

func (suite *ApiTestSuite) SetupSuite() {
	suite.Client = NewClient("localhost", 5001)
}

func (suite *ApiTestSuite) SetupTest() {
	suite.cleanup()
}

func (suite *ApiTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *ApiTestSuite) TestAllApps() {
	webapps, err := suite.Client.GetApps()
	suite.Nil(err)
	suite.Len(webapps, 0)

	create := 50
	idsCreated := []string{}
	for i := 0; i < create; i++ {
		name := "TESTWEBAPP" + strconv.Itoa(i)
		suite.Client.CreateApp(&types.App{ID: name})
		idsCreated = append(idsCreated, name)
	}
	webapps, err = suite.Client.GetApps()
	suite.Nil(err)
	suite.Len(webapps, create)
	idsRetrieved := []string{}
	for _, webapp := range webapps {
		idsRetrieved = append(idsRetrieved, webapp.ID)
	}
	sort.Strings(idsCreated)
	sort.Strings(idsRetrieved)
	suite.EqualValues(idsCreated, idsRetrieved)
}

func (suite *ApiTestSuite) TestCreateApp() {
	toCreate := &types.App{ID: "TESTWEBAPP"}
	err := suite.Client.CreateApp(toCreate)
	suite.Nil(err)
	overwrite := &types.App{ID: "TESTWEBAPP"}
	err = suite.Client.CreateApp(overwrite)
	suite.NotNil(err)

	webapp, err := suite.Client.GetApp("TESTWEBAPP")
	suite.Nil(err)
	suite.Equal(toCreate, webapp)
}

func (suite *ApiTestSuite) TestDelete() {
	toCreate := &types.App{ID: "TESTWEBAPP"}
	err := suite.Client.CreateApp(toCreate)
	suite.Nil(err)
	_, err = suite.Client.GetApp("TESTWEBAPP")
	suite.Nil(err)

	suite.Client.DeleteApp("TESTWEBAPP")
	_, err = suite.Client.GetApp("TESTWEBAPP")
	suite.NotNil(err)
}

func TestApiTestSuite(t *testing.T) {
	registry := modules.CreateAndRun("redis://localhost:6379", "testGroup1", "myMachine", "127.0.0.1", ":5002")
	suite.Run(t, &ApiTestSuite{runContext: registry.runContext})
	registry.Stop()
}
