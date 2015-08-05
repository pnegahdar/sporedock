package types

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestApp struct {
	name string
}

func (n TestApp) GetName() string {
	return n.name
}


type TypesTestSuite struct {
	suite.Suite
}

type Namable interface {
	GetName() string
}

var testMap = map[string]Namable{"test" : TestApp{}}

func (suite *TypesTestSuite) TestTypeMetaExtractor() {
	var app TestApp
	meta, err := NewMeta(app)
	suite.Nil(err)
	suite.Equal(TypeMeta{IsStruct: false, TypeName: "types.TestApp"}, meta)

	var appSlice []TestApp
	meta, err = NewMeta(appSlice)
	suite.Nil(err)
	suite.Equal(TypeMeta{IsStruct: true, TypeName: "types.TestApp"}, meta)

	appStructPointer := &TestApp{name: "TestApp"}
	meta, err = NewMeta(appStructPointer)
	suite.Nil(err)
	suite.Equal(TypeMeta{IsStruct: false, TypeName: "types.TestApp"}, meta)

	appSlicePointer := &[]TestApp{TestApp{name: "TestApp"}}
	meta, err = NewMeta(appSlicePointer)
	suite.Nil(err)
	suite.Equal(TypeMeta{IsStruct: true, TypeName: "types.TestApp"}, meta)
	// Todo(parham): Key namespace tests

	appInterface := testMap["test"]
	meta, err = NewMeta(appInterface)
	suite.Nil(err)
	suite.Equal(TypeMeta{IsStruct: false, TypeName: "types.TestApp"}, meta)
	meta, err = NewMeta(&appInterface)
	suite.Nil(err)
	suite.Equal(TypeMeta{IsStruct: false, TypeName: "types.TestApp"}, meta)
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}
