package grunts

import (
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"github.com/stretchr/testify/suite"
	"testing"
)

const testGroupName = "testGroup"

type TestTypeA struct {
	ID    string
	Extra string
}

type TestTypeB struct {
	ID    string
	Extra string
}

type GenericStoreTestSuite struct {
	suite.Suite
	registry *GruntRegistry
}

func (suite *GenericStoreTestSuite) cleanup() {
	typeA := &TestTypeA{}
	typeB := &TestTypeB{}
	err := suite.getStore().DeleteAll(typeA)
	suite.Nil(err)
	err = suite.getStore().DeleteAll(typeB)
	suite.Nil(err)
}

func (suite *GenericStoreTestSuite) SetupTest() {
	suite.cleanup()
}

func (suite *GenericStoreTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *GenericStoreTestSuite) getStore() types.SporeStore {
	return suite.registry.Context.Store
}

func (suite *GenericStoreTestSuite) TestGet() {
	retType := &TestTypeB{}
	err := suite.getStore().Get(retType, "ID_DNE")
	suite.Equal(err, types.ErrNoneFound)

	newSet := &TestTypeA{
		ID:    utils.GenGuid(),
		Extra: utils.GenGuid()}
	err = suite.getStore().Set(newSet, newSet.ID, types.SentinelEnd)
	suite.Nil(err)
	ret := &TestTypeA{}
	suite.getStore().Get(ret, newSet.ID)
	suite.Equal(ret.Extra, newSet.Extra)
}

func (suite *GenericStoreTestSuite) TestExists() {

	newSet := &TestTypeA{
		ID:    utils.GenGuid(),
		Extra: utils.GenGuid()}

	exists, err := suite.getStore().Exists(newSet, newSet.ID)
	suite.False(exists)
	suite.Nil(err)

	err = suite.getStore().Set(newSet, newSet.ID, types.SentinelEnd)
	suite.Nil(err)

	exists, err = suite.getStore().Exists(newSet, newSet.ID)
	suite.True(exists)
	suite.Nil(err)

}

func (suite *GenericStoreTestSuite) TestGetAll() {
	ids := []string{}
	var err error
	countInsert := 100
	for i := 0; i < countInsert; i++ {
		newSet := &TestTypeA{
			ID:    utils.GenGuid(),
			Extra: utils.GenGuid()}
		err = suite.getStore().Set(newSet, newSet.ID, types.SentinelEnd)
		ids = append(ids, newSet.ID)
		suite.Nil(err)
	}

	retAll := &[]TestTypeA{}
	err = suite.getStore().GetAll(retAll, 0, types.SentinelEnd)
	suite.Nil(err)
	suite.Equal(len(*retAll), countInsert)

	countRetrieve := 50
	retAll = &[]TestTypeA{}
	err = suite.getStore().GetAll(retAll, 0, countRetrieve)
	suite.Equal(len(*retAll), countRetrieve)

}

func (suite *GenericStoreTestSuite) TestSet() {
	id := utils.GenGuid()
	newSet := &TestTypeA{
		ID:    id,
		Extra: "TestExtraA"}
	err := suite.getStore().Set(newSet, newSet.ID, types.SentinelEnd)
	suite.Nil(err)
	retTest := &TestTypeA{}
	err = suite.getStore().Get(retTest, newSet.ID)
	suite.Nil(err)
	suite.Equal(retTest.Extra, newSet.Extra)

	newSet.Extra = "A"
	retTest = &TestTypeA{}
	err = suite.getStore().Set(newSet, newSet.ID, types.SentinelEnd)
	suite.Equal(err, types.ErrIDExists)

	newSetB := &TestTypeB{ID: id, Extra: "B"}
	err = suite.getStore().Set(newSetB, newSetB.ID, types.SentinelEnd)
	suite.Nil(err)
	retTest = &TestTypeA{}
	retTestB := &TestTypeB{}
	err = suite.getStore().Get(retTest, newSet.ID)
	suite.Nil(err)
	err = suite.getStore().Get(retTestB, newSetB.ID)
	suite.Nil(err)
	suite.Equal(retTest.Extra, "TestExtraA")
	suite.Equal(retTestB.Extra, "B")

}

func (suite *GenericStoreTestSuite) TestDelete() {
	first := TestTypeA{
		ID:    utils.GenGuid(),
		Extra: utils.GenGuid()}
	second := TestTypeA{
		ID:    utils.GenGuid(),
		Extra: utils.GenGuid()}
	err := suite.getStore().Set(first, first.ID, types.SentinelEnd)
	suite.Nil(err)
	err = suite.getStore().Set(second, second.ID, types.SentinelEnd)
	suite.Nil(err)

	exists, err := suite.getStore().Exists(second, second.ID)
	suite.True(exists)
	err = suite.getStore().Delete(second, second.ID)
	suite.Nil(err)
	exists, err = suite.getStore().Exists(second, second.ID)
	suite.False(exists)
	suite.Nil(err)
	exists, err = suite.getStore().Exists(first, first.ID)
	suite.True(exists)
	suite.Nil(err)

}

func (suite *GenericStoreTestSuite) TestDeleteAll() {
	first := TestTypeA{
		ID:    utils.GenGuid(),
		Extra: utils.GenGuid()}
	second := TestTypeA{
		ID:    utils.GenGuid(),
		Extra: utils.GenGuid()}

	err := suite.getStore().Set(first, first.ID, types.SentinelEnd)
	suite.Nil(err)
	err = suite.getStore().Set(second, second.ID, types.SentinelEnd)
	suite.Nil(err)

	err = suite.getStore().DeleteAll(second)
	suite.Nil(err)

	exists, err := suite.getStore().Exists(second, second.ID)
	suite.False(exists)
	suite.Nil(err)
	exists, err = suite.getStore().Exists(first, first.ID)
	suite.False(exists)
	suite.Nil(err)
}

func TestAllStores(t *testing.T) {
	storeTest := GenericStoreTestSuite{}
	storesToTest := []string{"redis://localhost:6379"}
	for _, store := range storesToTest {
		run := CreateAndRun(store, "testGroup", "myMachine", "127.0.0.1", ":5000")
		storeTest.registry = run
		suite.Run(t, &storeTest)
		run.Stop()
	}
}
