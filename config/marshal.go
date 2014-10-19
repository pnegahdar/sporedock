package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/aryann/difflib"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
	"strings"
)

func indentJSon(marshalled []byte) string {
	var buffer bytes.Buffer
	err := json.Indent(&buffer, marshalled, "", "    ")
	utils.HandleError(err)
	return buffer.String()
}

func GetClusterConfig() Cluster {
	return Cluster{}
}
func SetClusterConfig(cluster Cluster) {
	//	data, err := json.Marshal(cluster)
	//	utils.HandleError(err)
	//	data_str := string(data[:])
	//	_, err = etcd.EtcdClient.CreateInOrder(settings.ETCD_CONFIGS_KEY, data_str, 0)
	//	utils.HandleError(err)
	//	_, err = etcd.EtcdClient.Set(settings.ETCD_CURRENT_CONFIG_KEY, data_str, 0)
	ConvertClusterConfigToKeySet(cluster)
}
func ImportClusterConfigFromFile(filepath string) {
	var cluster Cluster
	fileData, err := ioutil.ReadFile(filepath)
	if err != nil {
		utils.HandleError(errors.New("Unable to find cluster config file specified."))
	}
	err = json.Unmarshal(fileData, &cluster)
	if err != nil {
		utils.HandleError(errors.New("Error JSON parsing the cluster config file specified."))
	}
	var noschema interface{}
	json.Unmarshal(fileData, &noschema)
	full_json_marshal, err := json.Marshal(noschema)
	utils.HandleError(err)
	detected_json_marshal, err := json.Marshal(cluster)
	utils.HandleError(err)
	if len(full_json_marshal) != len(detected_json_marshal) {
		diffs := difflib.Diff(strings.Split(indentJSon(full_json_marshal), "\n"), strings.Split(indentJSon(detected_json_marshal), "\n"))
		utils.LogWarn("Diff between detected and provided JSON:")
		for _, diff := range diffs {
			if diff.Delta != difflib.Common {
				utils.LogWarn(diff.Delta.String() + "     " + diff.Payload)
			}
		}
		utils.HandleError(errors.New("The JSON provided has bad structure."))
	}
	SetClusterConfig(cluster)
}
func ExportClusterToFile(filepath string) {

}
