package types

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/pnegahdar/sporedock/utils"
	"strings"
)

type DockerApp interface {
	DockerContainerOptions(runContext *RunContext, guid RunGuid) docker.CreateContainerOptions
	GetID() string
}

func dockerNameSpacePrefix(runContext *RunContext, extra ...string) string { return runContext.NamespacePrefix("-", extra...) }

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

func fullDockerAppName(guid RunGuid, containerList []docker.APIContainers) string {
	for _, cont := range containerList {
		for _, name := range cont.Names {
			if guid != "" && strings.Contains(name, string(guid)) {
				return name
			}
		}
	}
	return ""
}

func hasApp(guid RunGuid, containerList []docker.APIContainers) bool {
	containerName := fullDockerAppName(guid, containerList)
	return containerName != ""
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

func PullApp(runContext *RunContext, app DockerApp) {
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
	}
	utils.HandleError(err)
}

func createContainer(runContext *RunContext, guid RunGuid, app DockerApp) string {
	utils.LogInfoF("Creating container for %v.", app.GetID())
	containerConfig := app.DockerContainerOptions(runContext, guid)
	container, err := runContext.DockerClient.CreateContainer(containerConfig)
	utils.HandleError(err)
	return container.ID
}

func runApp(runContext *RunContext, containerID string, guid RunGuid, app DockerApp) {
	err := runContext.DockerClient.StartContainer(containerID, app.DockerContainerOptions(runContext, guid).HostConfig)
	utils.HandleError(err)
	EventDockerAppStart.EmitToSelf(runContext)
}

// Removes all the locations for nodes in the cluster that no longer exist.
func CleanDeadApps(runContext *RunContext) {
	listContainerOptions := docker.ListContainersOptions{All: true}
	containersAll, err := runContext.DockerClient.ListContainers(listContainerOptions)
	utils.HandleError(err)
	for _, cont := range containersAll {
		for _, name := range cont.Names {
			if strings.Contains(name, dockerNameSpacePrefix(runContext)) {
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

func CleanupRemovedApps(runContext *RunContext, guidsToKeep []RunGuid) {
	listContainerOptions := docker.ListContainersOptions{All: true}
	containersAll, err := runContext.DockerClient.ListContainers(listContainerOptions)
	utils.HandleError(err)
	for _, cont := range containersAll {
		delete := true
		for _, name := range cont.Names {
			if strings.Contains(name, dockerNameSpacePrefix(runContext)) {
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

func RunApp(runContext *RunContext, guid RunGuid, app *App) {
	containersAll, err := runContext.DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	utils.HandleError(err)
	containersRunning, err := runContext.DockerClient.ListContainers(docker.ListContainersOptions{All: false})
	utils.HandleError(err)
	if hasApp(guid, containersRunning) {
		return
	}
	if hasApp(guid, containersAll) {
		return
	}
	containerID := createContainer(runContext, guid, app)
	utils.LogInfoF("Running app. App Guid: %v App ID: %v Container ID: %v", guid, app.ID, containerID)
	runApp(runContext, containerID, guid, app)
	utils.LogInfoF("Started app. App Guid: %v App ID: %v Container ID: %v", guid, app.ID, containerID)

}
