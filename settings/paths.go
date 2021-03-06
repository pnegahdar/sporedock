package settings

import (
	"github.com/mitchellh/go-homedir"
	"os"
	"path"
)

var ProjectBaseDir, _ = homedir.Expand(path.Join("~/.config/", AppName))

func GetProjectPath(addon_path string) string {
	return path.Join(ProjectBaseDir, addon_path)
}

func GetInstanceIdConfPath() string {
	return GetProjectPath("INSTNACE_ID")
}

func GetDiscoveryConfPath() string {
	return GetProjectPath("DISCOVERY")
}

func GetRaftDataDir() string {
	return GetProjectPath("raft/")
}

func init() {
	os.Mkdir(ProjectBaseDir, 0700)
}
