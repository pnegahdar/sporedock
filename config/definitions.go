package config

import (
	"encoding/json"
	"errors"
	"github.com/aryann/difflib"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
	"strings"
)

type Env struct {
	Env map[string]string `flatten:"{{ .ID }}/{{ .KEY }}"`
	ID  string            `flatten:"{{ .ID }}/"`
}

type WebApp struct {
	EnvIDs         []string `flatten:"{{ .ID }}/EnvsIDs/"`
	ID             string   `flatten:"{{ .ID }}"`
	Image          string   `flatten:"{{ .ID }}/Image"`
	InternalPort   string   `flatten:"{{ .ID }}/InternalPort"`
	StartupCommand string   `flatten:"{{ .ID }}/StartupCommand"`
	WebEndpoints   []string `flatten:"{{ .ID }}/WebEndpoints/"`
}

type Cluster struct {
	Envs    []Env    `flatten:"/sporedock/clusters/{{ .ID }}/envs/"`
	ID      string   `flatten:"/sporedock/clusters/{{ .ID }}/"`
	WebApps []WebApp `flatten:"/sporedock/cluster/{{ .ID }}/webapps/"`
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
	Flatten(c)
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
		diffs := difflib.Diff(strings.Split(indentJSon(full_json_marshal), "\n"), strings.Split(indentJSon([]byte(detected_json_marshal)), "\n"))
		utils.LogWarn("Diff between detected and provided JSON:")
		for _, diff := range diffs {
			if diff.Delta != difflib.Common {
				utils.LogWarn(diff.Delta.String() + "     " + diff.Payload)
			}
		}
		utils.HandleError(errors.New("The JSON provided has bad structure."))
	}
	&c.Validate()
}

func (c Cluster) EtcdSet() {
	c.Validate()
	current_config, err := server.EtcdClient().Get(ETCD_CURRENT_CONFIG_KEY, false, false)
	utils.HandleError(err)
	// Cache existing key
	if current_config != nil {
		_, err1 := server.EtcdClient().CreateInOrder(ETCD_CONFIGS_KEY, current_config.Action, 0)
		utils.HandleError(err1)
	}
	cluster_json, err := c.Marshall()
	utils.HandleError(err)
	server.EtcdClient().Set(ETCD_CURRENT_CONFIG_KEY, cluster_json)
}

func (c *Cluster) EtcdGet() {
	current_config, err := server.EtcdClient().Get(ETCD_CURRENT_CONFIG_KEY, false, false)
	utils.HandleError(err)
	c.UnMarshall(current_config.Action)
}
