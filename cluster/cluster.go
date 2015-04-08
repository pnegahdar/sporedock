package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aryann/difflib"
	"github.com/coreos/etcd/discovery"
	"github.com/pnegahdar/sporedock/discovery"
	"github.com/pnegahdar/sporedock/utils"
	"github.com/samalba/dockerclient"
	"io/ioutil"
	"strings"
)

const clusterLogLength = 100
const clusterconfigKey = "sporedock:cluster:current"

type DockerAppIter interface {
	IterApps() []DockerApp
}

type Cluster struct {
	Envs       Envs
	ID         string
	WebApps    WebApps
	WorkerApps WorkerApps
}

type Cluster struct {
	Envs Envs
}

func (c Cluster) IterApps() []DockerApp {
	apps := []DockerApp{}
	for _, app := range c.WebApps {
		apps = append(apps, app)
	}
	for _, app := range c.WorkerApps {
		apps = append(apps, app)
	}
	return apps
}

func (c Cluster) GetEnv(envID string) Env {
	for _, env := range c.Envs {
		if env.ID == envID {
			return env
		}
	}
	utils.HandleError(errors.New(fmt.Sprintf("Error: Env %v not found in cluster.", envID)))
	return Env{}
}

func (c Cluster) Validate() {
	flattenCluster(c)
}

func (c Cluster) Export(filepath string) {
	marshalled, err := marshall(c)
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
	detected_json_marshal, err := marshall(c)
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
	c.Push()
}

func (c Cluster) Push() {
	store := store.GetStore()
	c.Validate()
	cluster_json, err := marshall(c)
	utils.HandleError(err)
	err = store.SetKeyWithLog(clusterconfigKey, cluster_json, clusterLogLength)
	utils.HandleError(err)
}

func (c *Cluster) Pull() {
	store := store.GetStore()
	resp, err := store.GetKey(clusterconfigKey)
	utils.HandleError(err)
	unmarshall(resp, c)
}

func GetCurrentCluster() Cluster {
	c := Cluster{}
	c.Pull()
	return c
}
