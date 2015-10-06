package types

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Module interface {
	ProcName() string
	ShouldRun(runContext *RunContext) bool
	Init(runContext *RunContext)
	Run(runContext *RunContext)
	Stop()
}

type CliModule interface {
	InitCli(runContext *RunContext)
}

type Identifiable interface {
	GetID() string
}

type Creatable interface {
	Identifiable
	Validate(*RunContext) error
}

type Validable interface {
	Identifiable
	Validate(*RunContext) error
}

const CpuMemMultiplier = 1 / 512

type Sizable struct {
	Cpus float64
	Mem  float64
}

func GetSize(cpu float64, mem float64) float64 {
	return cpu + (mem * CpuMemMultiplier)
}

type TypeMeta struct {
	IsSlice  bool
	TypeName string
}

func NewMeta(v interface{}) (TypeMeta, error) {
	typeof := reflect.TypeOf(v)
	kind := typeof.Kind()
	if kind == reflect.Ptr {
		typeof = reflect.ValueOf(v).Elem().Type()
		kind = typeof.Kind()
	}
	var isSlice bool
	var typeName string
	switch kind {
	case reflect.Slice:
		isSlice = true
		typeName = fmt.Sprint(typeof.Elem())
	case reflect.Struct:
		typeName = fmt.Sprint(typeof)
		isSlice = false
	case reflect.Interface:
		isSlice = false
		typeName = fmt.Sprint(reflect.ValueOf(v).Elem().Elem().Type())
	case reflect.String, reflect.Int, reflect.Bool:
		isSlice = false
		typeName = fmt.Sprint(typeof)
	default:
		err := errors.New("Type not struct or slice")
		return TypeMeta{}, err
	}
	return TypeMeta{IsSlice: isSlice, TypeName: strings.TrimPrefix(typeName, "*")}, nil
}
