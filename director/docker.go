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
		utils.LogWarn(fmt.Sprintf("App %v already ran but exited. Not restarting.", app.GetName()))
		waitGroup.Done()
		return
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

func CleanupRemovedApps(appsToKeep []string) {
	dc := CachedDockerClient()
	resp, err := dc.ListContainers(true)
	utils.HandleError(err)
	for _, cont := range resp {
		for _, name := range cont.Names {
			if strings.HasPrefix(name, "/Sporedock") {
				if !In(appsToKeep, name) {
					dc.RemoveContainer(name, true)
				}
			}
		}

	}
}

func pathLastPart(path string) string {
	allPaths := strings.Split(path, "/")
	return allPaths[len(allPaths)-1]
}

// Removes all the locations for nodes in the cluster that no longer exist.
func CleanupLocations() {
	currentCluster := cluster.GetCurrentCluster()
	currentManifest := cluster.GetCurrentManifest()
	machineList := []string{}
	for _, machine := range currentManifest {
		machineList = append(machineList, machine.Machine.Name)
	}
	// Remove APPS DNE
	appNames := []string{}
	for _, app := range currentCluster.IterApps() {
		appNames = append(appNames, app.GetName())
	}
	resp, err := server.EtcdClient().Get(cluster.AppLocationsDirKey, true, false)
	if err != nil && strings.Index(err.Error(), "Key not found") != -1 {
		return
	}
	utils.HandleError(err)
	for _, node := range resp.Node.Nodes {
		appName := pathLastPart(node.Key)
		if !In(appNames, appName) {
			utils.LogDebug(fmt.Sprintf("App %v no longer exists removing loc.", appName))
			_, err := server.EtcdClient().Delete(node.Key, true)
			utils.HandleError(err)
		}
	}
	// Remove Machines DNE
	for _, app := range currentCluster.IterApps() {
		keyName := cluster.AppLocationsDirKey + app.GetName() + "/"
		resp, err := server.EtcdClient().Get(keyName, true, false)
		if err != nil && strings.Index(err.Error(), "Key not found") != -1 {
			continue
		}
		utils.HandleError(err)
		for _, node := range resp.Node.Nodes {
			machineName := pathLastPart(node.Key)
			if !In(machineList, machineName) {
				utils.LogDebug(fmt.Sprintf("Machine %v no longer exists removing loc.", machineName))
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
		keyName := cluster.AppLocationsDirKey + appName + "/" + machine.Name
		resp, err := dc.InspectContainer(appName)
		utils.HandleError(err)
		// Remove dead app
		if !resp.State.Running {
			_, err := server.EtcdClient().Delete(keyName, true)
			if err != nil && strings.Index(err.Error(), "Key not found") != -1 {
				continue
			}
			utils.LogDebug(fmt.Sprintf("Removed dead app location %v", appName))
			utils.HandleError(err)
			continue
		}
		bindings := resp.NetworkSettings.Ports
		for k, v := range bindings {
			if k == "80/tcp" {
				//Todo(parham): Only allows for one per node

				_, err := server.EtcdClient().Set(keyName, v[0].HostPort, 0)
				utils.HandleError(err)
			}
		}
	}
}
func CleanDeadApps() {
	dc := CachedDockerClient()
	resp, err := dc.ListContainers(true)
	utils.HandleError(err)
	for _, cont := range resp {
		for _, name := range cont.Names {
			utils.LogDebug(fmt.Sprintf("%v", name))
			if strings.HasPrefix(name, "/Sporedock") {
				resp, err := dc.InspectContainer(name)
				utils.HandleError(err)
				if !resp.State.Running {
					utils.LogDebug(fmt.Sprintf("App %v looks dead removing.", name))
					dc.RemoveContainer(name, true)
				}
			}
		}

	}
}
