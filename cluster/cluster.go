package cluster

import (
	"encoding/json"
	"errors"
	"github.com/aryann/difflib"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
	"strings"
)

type Cluster struct {
	Envs       Envs       `flatten:"/sporedock/clusters/{{ .ID }}/Envs/"`
	ID         string     `flatten:"/sporedock/clusters/{{ .ID }}/"`
	WebApps    WebApps    `flatten:"/sporedock/cluster/{{ .ID }}/WebApps/"`
	WorkerApps WorkerApps `flatten:"/sporedock/cluster/{{ .ID }}/WorkerApps/"`
}

func (c Cluster) Marshall() (string, error) {
	resp, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(resp[:]), nil
}

func (c *Cluster) UnMarshall(data string) error {
	err := json.Unmarshal([]byte(data), c)
	if err != nil {
		return err
	}
	return nil
}

func (c Cluster) Validate() {
	flattenCluster(c)
}

func (c Cluster) Export(filepath string) {
	marshalled, err := c.Marshall()
	utils.HandleError(err)
	ioutil.WriteFile(filepath, []byte(marshalled), 700)
}

func (c *Cluster) Import(filepath string) {
	fileData, err := ioutil.ReadFile(filepath)
	if err != nil {
		utils.HandleError(errors.New("Unable to find cluster config file specified."))
	}
	err = json.Unmarshal(fileData, c)
	if err != nil {
		utils.HandleError(errors.New("Error JSON parsing the cluster config file specified."))
	}
	var noschema interface{}
	json.Unmarshal(fileData, &noschema)
	full_json_marshal, err := json.Marshal(noschema)
	utils.HandleError(err)
	detected_json_marshal, err := c.Marshall()
	utils.HandleError(err)
	if len(full_json_marshal) != len(detected_json_marshal) {
		diffs := difflib.Diff(strings.Split(indentJSon(full_json_marshal), "\n"), strings.Split(
			indentJSon([]byte(detected_json_marshal)), "\n"))
		utils.LogWarn("Diff between detected and provided JSON:")
		for _, diff := range diffs {
			if diff.Delta != difflib.Common {
				utils.LogWarn(diff.Delta.String() + "     " + diff.Payload)
			}
		}
		utils.HandleError(errors.New("The JSON provided has bad structure."))
	}
	c.Validate()
	c.Set()
}

func (c Cluster) Set() {
	c.Validate()
	cluster_json, err := c.Marshall()
	utils.HandleError(err)
	_, err1 := server.EtcdClient().CreateInOrder(ConfigsKey, cluster_json, 0)
	utils.HandleError(err1)
	_, err2 := server.EtcdClient().Set(CurrentConfigKey, cluster_json, 0)
	utils.HandleError(err2)
}

func (c *Cluster) Get() {
	current_config, err := server.EtcdClient().Get(CurrentConfigKey, false, false)
	utils.HandleError(err)
	c.UnMarshall(current_config.Action)
}
