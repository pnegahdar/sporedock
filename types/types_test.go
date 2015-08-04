package types

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestApp struct {
	name string
}

type TypesTestSuite struct {
	suite.Suite
}

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
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}
