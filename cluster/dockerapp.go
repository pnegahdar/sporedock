package cluster

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/pnegahdar/sporedock/types"
	"github.com/pnegahdar/sporedock/utils"
	"strings"
)

type DockerApp interface {
	DockerContainerOptions(runContext *types.RunContext, guid RunGuid) docker.CreateContainerOptions
	GetID() string
}

func nameSpacePrefix(runContext *types.RunContext) string { return runContext.NamespacePrefix("") }

func hasApp(guid RunGuid, containerList []docker.APIContainers) bool {
	for _, cont := range containerList {
		for _, name := range cont.Names {
			if guid != "" && strings.Contains(name, string(guid)) {
				return true
			}
		}
	}
	return false
}

func hasImage(imagelist []docker.APIImages, fullImageName string) bool {
	for _, im := range imagelist {
		for _, ims := range im.RepoTags {
			if fullImageName == ims {
				return true
			}
		}
	}
	return false
}

type parsedImage struct {
	Tag  string
	Name string
	Full string
}

func parseDockerImage(image string) (parsedImage, error) {
	parts := strings.Split(image, ":")
	if len(parts) == 1 {
		parts = append(parts, "latest")
	}
	imgMeta := parsedImage{Tag: parts[1], Name: parts[1], Full: strings.Join(parts, ":")}
	return imgMeta, nil
}

func PullApp(runContext *types.RunContext, app DockerApp) {
	imgs, err := runContext.DockerClient.ListImages(docker.ListImagesOptions{All: true})
	utils.HandleError(err)
	parsedImage, err := parseDockerImage(app.DockerContainerOptions(runContext, "").Config.Image)
	utils.HandleError(err)
	// Todo(parham): notify on change here...
	if !hasImage(imgs, parsedImage.Full) {
		utils.LogInfoF("Started pulling docker image %v", parsedImage.Full)
		imageOptions := docker.PullImageOptions{Repository: parsedImage.Name, Tag: parsedImage.Tag}
		authConfig := docker.AuthConfiguration{}
		err = runContext.DockerClient.PullImage(imageOptions, authConfig)
		utils.HandleError(err)
		utils.LogInfoF("Finished pulling docker image %v", parsedImage.Full)
	} else {
		utils.LogInfoF("Already have image %v", parsedImage.Full)
	}
	utils.HandleError(err)
}

func pathLastPart(path string) string {
	allPaths := strings.Split(path, "/")
	return allPaths[len(allPaths)-1]
}

func createContainer(runContext *types.RunContext, guid RunGuid, app DockerApp) string {
	utils.LogInfoF("Creating container for %v.", app.GetID())
	containerConfig := app.DockerContainerOptions(runContext, guid)
	container, err := runContext.DockerClient.CreateContainer(containerConfig)
	utils.HandleError(err)
	return container.ID
}

func runApp(runContext *types.RunContext, containerID string, guid RunGuid, app DockerApp) {
	utils.LogInfoF("Running app %v of %v", containerID, app.GetID())
	err := runContext.DockerClient.StartContainer(containerID, app.DockerContainerOptions(runContext, guid).HostConfig)
	utils.HandleError(err)
}

// Removes all the locations for nodes in the cluster that no longer exist.
func CleanDeadApps(runContext *types.RunContext) {
	listContainerOptions := docker.ListContainersOptions{All: true}
	containersAll, err := runContext.DockerClient.ListContainers(listContainerOptions)
	utils.HandleError(err)
	for _, cont := range containersAll {
		for _, name := range cont.Names {
			if strings.Contains(name, nameSpacePrefix(runContext)) {
				resp, err := runContext.DockerClient.InspectContainer(name)
				utils.HandleError(err)
				if !resp.State.Running {
					utils.LogInfoF("App %v looks dead removing.", name)
					name = name[1:len(name)]
					removeOptions := docker.RemoveContainerOptions{ID: name, Force: true}
					err := runContext.DockerClient.RemoveContainer(removeOptions)
					utils.HandleError(err)
				}
			}
		}

	}
}
func CleanupRemovedApps(runContext *types.RunContext, guidsToKeep []RunGuid) {
	listContainerOptions := docker.ListContainersOptions{All: true}
	containersAll, err := runContext.DockerClient.ListContainers(listContainerOptions)
	utils.HandleError(err)
	for _, cont := range containersAll {
		delete := true
		for _, name := range cont.Names {
			if strings.Contains(name, nameSpacePrefix(runContext)) {
				for _, guid := range guidsToKeep {
					if strings.Contains(name, string(guid)) {
						delete = false
					}
				}
			}
		}
		if delete {
			runContext.DockerClient.RemoveContainer(docker.RemoveContainerOptions{ID: cont.ID, Force: true})
		}
	}
}
func RunApp(runContext *types.RunContext, guid RunGuid, app *App) {
	containersAll, err := runContext.DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	utils.HandleError(err)
	containersRunning, err := runContext.DockerClient.ListContainers(docker.ListContainersOptions{All: false})
	utils.HandleError(err)
	appRunning := hasApp(guid, containersRunning)
	appExists := hasApp(guid, containersAll)
	if appRunning {
		return
	}
	if appExists {
		return
	}
	containerID := createContainer(runContext, guid, app)
	utils.LogInfoF("Running app. App Guid: %v App ID: %v Container ID: %v", guid, app.ID, containerID)
	runApp(runContext, containerID, guid, app)
	utils.LogInfoF("Started app. App Guid: %v App ID: %v Container ID: %v", guid, app.ID, containerID)

}
