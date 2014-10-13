package manifest

import (
	"encoding/json"
	"errors"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
)

func GetClusterConfig() Cluster {
	return Cluster{}
}
func SetClusterConfig(Cluster) {
	// Assert primary first.

}
func ImportClusterConfigFromFile(filepath string) Cluster {
	var cluster Cluster
	fileData, err := ioutil.ReadFile(filepath)
	if err != nil {
		utils.HandleError(errors.New("Unable to find cluster config file specified."))
	}
	err = json.Unmarshal(fileData, &cluster)
	if err != nil {
		utils.HandleError(errors.New("Error JSON parsing the cluster config file specified."))
	}
	return cluster
}
func ExportClusterToFile(filepath string) {

}
