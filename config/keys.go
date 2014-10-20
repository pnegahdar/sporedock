package config

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/pnegahdar/sporedock/utils"
	"reflect"
	"text/template"
)

func flatten(prefix string, value reflect.Value, data map[string]string) {
	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			var buffer bytes.Buffer
			type_key := value.Type().Field(i)
			item := value.Field(i)
			tags := type_key.Tag
			etcd_tag := tags.Get("etcd")
			tmpl, _ := template.New("etcd").Parse(etcd_tag)
			tmpl.Execute(&buffer, value.Interface())
			flatten(prefix+buffer.String(), item, data)
		}
	case reflect.Slice:
		fmt.Println(value.Type())
		switch value.Type() {
		case reflect.Type([]string):
			for i := 0; i < value.Len(); i++ {
				add := value.Index(i).Interface().(string)
				data[prefix+add] = add
			}
		default:
			for i := 0; i < value.Len(); i++ {
				flatten(prefix, value.Index(i), data)
			}
		}
	case reflect.String:
		if _, exists := data[prefix]; exists {
			utils.HandleError(errors.New("Duplicate key " + prefix))
		}
		data[prefix] = value.Interface().(string)
	case reflect.Map:
		for k, v := range value.Interface().(map[string]string) {
			if _, exists := data[prefix+k]; exists {
				utils.HandleError(errors.New("Duplicate key " + prefix + k))
			}
			data[prefix+k] = v
		}
	default:
		utils.HandleError(errors.New("Unidentified type slipped though. Please check."))
	}
}

func ConvertClusterConfigToKeySet(cluster Cluster) map[string]string {
	val := reflect.ValueOf(cluster)
	var data = make(map[string]string)
	flatten("", val, data)
	//	for k, v := range(data){
	//		fmt.Println(k,"    ->   ", v)
	//	}
	return data
}
