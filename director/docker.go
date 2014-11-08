package director

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/pnegahdar/sporedock/utils"
	"github.com/samalba/dockerclient"
	"io/ioutil"
	"os"
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

func PullApp(image string, tag string) {
	client := CachedDockerClient()
	err := client.PullImage(image, tag)
	utils.HandleError(err)
}
