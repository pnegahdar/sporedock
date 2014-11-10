package director

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/pnegahdar/sporedock/cluster"
	"github.com/pnegahdar/sporedock/discovery"
	"github.com/pnegahdar/sporedock/utils"
	"github.com/samalba/dockerclient"
	"io/ioutil"
	"os"
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
	image := app.GetImage()
	tag := app.GetTag()
	utils.HandleError(err)
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

func hasApp(appId string, containers []dockerclient.Container) bool {
	for _, cont := range containers {
		for _, name := range cont.Names {
			if appId == name {
				return true
			}
		}
	}
	return false
}

func runDocker(app cluster.DockerApp) {
	client := CachedDockerClient()
	config := app.ContainerConfig()
	client.CreateContainer(&config, app.GetName())
}

func RunMyApps() {
	client := CachedDockerClient()
	containers, err := client.ListContainers(true)
	utils.HandleError(err)
	currentManifest := cluster.Manifests{}
	currentManifest.Pull()
	myManifest := currentManifest.MyManifest(discovery.CurrentMachine())
	apps := myManifest.IterApps()
	for _, app := range apps {
		if !hasApp(app.GetImage(), containers) {
			runDocker(app)
		}
	}
}
