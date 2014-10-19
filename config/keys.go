package config

import (
	"fmt"
	"reflect"
)

func convertValue(prefix string, val reflect.Value) {
	subjectType := reflect.TypeOf(val)
	for i := 0; i < subjectType.NumField(); i++ {
		field := subjectType.Field(i)
		name := subjectType.Field(i).Name
		tag := subjectType.Field(i).Tag.Get("etcd")
		fmt.Println(field, name, tag)
	}
}

func convertCluster(cluster Cluster) {
	value := reflect.ValueOf(cluster)
	convertValue(cluster.ID, value)
}

func ConvertClusterConfigToKeySet(cluster Cluster) {
	convertCluster(cluster)
}
