package config

import (
	"bytes"
	"errors"
	"github.com/pnegahdar/sporedock/utils"
	"reflect"
	"text/template"
)

func flatten(prefix string, value reflect.Value, data *[][]string) {
	// ATTEMPT TO DO THIS AUTOMAGICALLY USING TAGS... SET ASIDE FOR NOW....
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
		for i := 0; i < value.Len(); i++ {
			flatten(prefix, value.Index(i), data)
		}
	case reflect.String:
		*data = append(*data, []string{prefix, value.Interface().(string)})
	case reflect.Map:
		for k, v := range value.Interface().(map[string]string) {
			*data = append(*data, []string{prefix + k, v})
		}
	default:
		utils.HandleError(errors.New("Unidentified type slipped though. Please check."))
	}
}

func ConvertClusterConfigToKeySet(cluster Cluster) [][]string {
	val := reflect.ValueOf(cluster)
	var data [][]string
	flatten("", val, &data)
	return data
}
