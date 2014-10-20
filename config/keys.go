package config

import (
	"bytes"
	"github.com/pnegahdar/sporedock/utils"
	"reflect"
	"text/template"
	"fmt"
)

func blah(prefix string, value reflect.Value ) [][]string {
	// ATTEMPT TO DO THIS AUTOMAGICALLY USING TAGS... SET ASIDE FOR NOW....
	var data [][]string
	for i := 0; i < value.NumField(); i++ {
		var buffer bytes.Buffer

		type_key := value.Type().Field(i)
		item := value.Field(i)
		tags := type_key.Tag
		etcd_tag := tags.Get("etcd")


		tmpl, err := template.New("etcd").Parse(etcd_tag)
		utils.HandleError(err)
		tmpl.Execute(&buffer, value.Interface())
		switch item.Kind() {
		case reflect.String:
			data = append(data, []string{buffer.String(), item.Interface().(string)})
		case reflect.Slice:
			for ind:=0; ind<item.Len();ind++{
				fmt.Println(item.Index(ind))
				data = append(data, flattenStruct(etcd_tag, item.Index(ind))...)
				fmt.Println(item.Index(ind).Interface())
		}
		}
	}
	fmt.Println(data)
	return data
}


func flattenStruct(prefix string, value reflect.Value ) [][]string {
	// ATTEMPT TO DO THIS AUTOMAGICALLY USING TAGS... SET ASIDE FOR NOW....
	var data [][]string
	switch value.Kind(){
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			var buffer bytes.Buffer
			type_key := value.Type().Field(i)
			item := value.Field(i)
			tags := type_key.Tag
			etcd_tag := tags.Get("etcd")
			tmpl, _ := template.New("etcd").Parse(etcd_tag)
			tmpl.Execute(&buffer, value.Interface())
			data = append(data, flattenStruct(buffer.String(), item)...)

		}
	case reflect.String:
		data = append(data, []string{prefix, value.Interface().(string)})
	}
	fmt.Println(data)
	return data
}

func ConvertClusterConfigToKeySet(cluster Cluster) {
	val := reflect.ValueOf(cluster)
	_ = flattenStruct("", val)
}
