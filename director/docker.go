package director

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/discovery"
	"github.com/pnegahdar/sporedock/server"
	"github.com/pnegahdar/sporedock/utils"
	"github.com/samalba/dockerclient"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type DockerAppConfig struct{}

type DockerApp interface {
	GetConfig() DockerAppConfig
}

const DockerHost = "https://192.168.59.103:2376"

var cachedDockerClient dockerclient.DockerClient

func CachedDockerClient() dockerclient.DockerClient {
	if cachedDockerClient == nil {
		tlsConfig := tls.Config{}
		if os.Getenv("DOCKER_TLS_VERIFY") == "1" {
			certDir := os.Getenv("DOCKER_CERT_PATH") + "/"
			certPath := certDir + "cert.pem"
			caPath := certDir + "ca.pem"
			keyPath := certDir + "key.pem"
			certPool := x509.NewCertPool()
			caFile, err := ioutil.ReadFile(caPath)
			utils.HandleError(err)
			certPool.AppendCertsFromPEM(caFile)
			cert, err := tls.LoadX509KeyPair(certPath, keyPath)
			utils.HandleError(err)
			tlsConfig.Certificates = []tls.Certificate{cert}
			tlsConfig.RootCAs = certPool
		}
		client, err := dockerclient.NewDockerClient(DockerHost, &tlsConfig)
		utils.HandleError(err)
		cachedDockerClient = client
	}
	return cachedDockerClient
}

func hasImage(imagelist []*dockerclient.Image, image, tag string) bool {
	for _, im := range imagelist {
		for _, ims := range im.RepoTags {
			if fmt.Sprintf("%v:%v", image, tag) == ims {
				return true
			}
		}
	}
	return false
}

func pullApp(app cluster.DockerApp, wg *sync.WaitGroup) {
	client := CachedDockerClient()
	imgs, err := client.ListImages()
	utils.HandleError(err)
	image := app.GetImage()
	tag := app.GetTag()
	// Todo(parham): notify on change here...
	if !hasImage(imgs, image, tag) {
		utils.LogDebug(fmt.Sprintf("Started pulling docker image %v:%v", image, tag))
		err = client.PullImage(image, tag)
		utils.LogInfo(fmt.Sprintf("Finished pulling docker image %v:%v", image, tag))
	} else {
		utils.LogDebug(fmt.Sprintf("Already have image %v:%v", image, tag))
	}
	utils.HandleError(err)
	wg.Done()
}

func RunAppSafe(app cluster.DockerApp, manifest cluster.MachineManifest, waitGroup *sync.WaitGroup) {
	dc := CachedDockerClient()
	containersAll, err := dc.ListContainers(true)
	utils.HandleError(err)
	containersRunning, err1 := dc.ListContainers(false)
	utils.HandleError(err1)
	idAppRunning := hasApp(app, containersRunning)
	idAppExists := hasApp(app, containersAll)
	if idAppRunning != "" {
		utils.LogDebug(fmt.Sprintf("App %v already running. Skipping.", app.GetName()))
		waitGroup.Done()
		return
	}
	if idAppExists != "" {
		utils.LogDebug(fmt.Sprintf("App %v already exists. Removing.", app.GetName()))
		dc.RemoveContainer(idAppExists, false)
	}
	id := createContainer(app)
	runApp(app, id)
	waitGroup.Done()
}
func hasApp(app cluster.DockerApp, containerList []dockerclient.Container) string {
	for _, cont := range containerList {
		for _, name := range cont.Names {
			if "/"+app.GetName() == name {
				return cont.Id
			}
		}
	}
	return ""
}

func createContainer(app cluster.DockerApp) string {
	utils.LogDebug(fmt.Sprintf("Creating container for %v.", app.GetName()))
	client := CachedDockerClient()
	config := app.ContainerConfig()
	id, err := client.CreateContainer(&config, app.GetName())
	utils.HandleError(err)
	return id
}

func runApp(app cluster.DockerApp, containerID string) {
	utils.LogDebug(fmt.Sprintf("Running app %v", app.GetName()))
	client := CachedDockerClient()
	hostConfig := app.HostConfig()
	err := client.StartContainer(containerID, &hostConfig)
	utils.HandleError(err)
}
func In(haystack []string, needle string) bool {
	for _, hay := range haystack {
		if hay == needle {
			return true
		}
	}
	return false
}

func CleanupApps(appsToKeep []string) {
	dc := CachedDockerClient()
	resp, err := dc.ListContainers(true)
	utils.HandleError(err)
	for _, cont := range resp {
		for _, name := range cont.Names {
			if strings.HasPrefix(name, "Sporedock") {
				if !In(appsToKeep, name) {
					dc.RemoveContainer(name, true)
				}
			}
		}

	}
}

// Removes all the locations for nodes in the cluster that no longer exist.
func CleanupLocations() {
	currentCluster := cluster.GetCurrentCluster()
	currentManifest := cluster.GetCurrentManifest()
	machineList := []string{}
	for _, machine := range currentManifest {
		machineList = append(machineList, machine.Machine.Name)
	}
	for _, app := range currentCluster.IterApps() {
		keyName := cluster.AppLocationsDirKey + app.GetName() + "/"
		resp, err := server.EtcdClient().Get(keyName, true, false)
		if err != nil && strings.Index(err.Error(), "Key not found") != -1 {
			continue
		}
		fmt.Println(resp)
		utils.HandleError(err)
		for _, node := range resp.Node.Nodes {
			if !In(machineList, node.Key) {
				_, err := server.EtcdClient().Delete(node.Key, true)
				utils.HandleError(err)
			}
		}
	}

}

func UpdateLocations(appNames []string) {
	dc := CachedDockerClient()
	machine := discovery.CurrentMachine()
	for _, appName := range appNames {
		resp, err := dc.InspectContainer(appName)
		utils.HandleError(err)
		bindings := resp.HostConfig.PortBindings
		for k, v := range bindings {
			if k == "80/tcp" {
				//Todo(parham): Only allows for one per node
				keyName := cluster.AppLocationsDirKey + appName + "/" + machine.Name
				_, err := server.EtcdClient().Set(keyName, v[0].HostPort, 0)
				utils.HandleError(err)
			}
		}

	}

}
