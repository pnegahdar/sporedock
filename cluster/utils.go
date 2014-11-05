package cluster

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/pnegahdar/sporedock/utils"
	"reflect"
	"text/template"
)

func addSafe(data map[string]string, key string, value string) {
	if _, exists := data[key]; exists {
		utils.HandleError(errors.New("Duplicate key " + key))
	}
	data[key] = value
}

func flatten(prefix string, value reflect.Value, data map[string]string) {
	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			var buffer bytes.Buffer
			flatten_tag := value.Type().Field(i).Tag.Get("flatten")
			tmpl, _ := template.New("flatten").Parse(flatten_tag)
			tmpl.Execute(&buffer, value.Interface())
			flatten(prefix+buffer.String(), value.Field(i), data)
		}
	case reflect.Slice:
		switch value.Type() {
		case reflect.TypeOf([]string{}):
			for i := 0; i < value.Len(); i++ {
				add := value.Index(i).Interface().(string)
				addSafe(data, prefix+add, add)
			}
		default:
			for i := 0; i < value.Len(); i++ {
				flatten(prefix, value.Index(i), data)
			}
		}
	case reflect.String:
		addSafe(data, prefix, value.Interface().(string))
	case reflect.Map:
		for k, v := range value.Interface().(map[string]string) {
			addSafe(data, prefix+k, v)
		}
	default:
		utils.HandleError(errors.New("Unidentified type slipped though. Please check."))
	}
}

func indentJSon(marshalled []byte) string {
	var buffer bytes.Buffer
	err := json.Indent(&buffer, marshalled, "", "    ")
	utils.HandleError(err)
	return buffer.String()
}

func flattenCluster(cluster Cluster) map[string]string {
	val := reflect.ValueOf(cluster)
	var data = make(map[string]string)
	flatten("", val, data)
	return data
}

func marshall(i interface{}) (string, error) {
	resp, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(resp[:]), nil
}

func unmarshall(data string, i *interface{}) error {
	err := json.Unmarshal([]byte(data), i)
	if err != nil {
		return err
	}
	return nil
}
