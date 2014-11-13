package settings

import (
	"code.google.com/p/go-uuid/uuid"
	"errors"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
)

func GetInstanceName() string {
	filePath := GetInstanceIdConfPath()
	data := getFileContentsString(filePath)
	if data == "" {
		// Todo(parham): use a non random uuid
		uuidBase := uuid.NewRandom()
		uuidString := uuid.NewSHA1(uuidBase, nil).String()
		writeFileContentsString(filePath, uuidString)
		data = uuidString
	}
	return data
}
func GetDiscoveryString() string {
	filePath := GetDiscoveryConfPath()
	content := getFileContentsString(filePath)
	if content == "" {
		utils.HandleError(errors.New("Must set discovery URI first with 'init' command"))
	}
	return content
}
func SetDiscoveryString(discovery string) {
	filePath := GetDiscoveryConfPath()
	writeFileContentsString(filePath, discovery)
}

func getFileContentsString(path string) string {
	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		utils.LogWarn(err.Error())
		return ""
	}
	return string(fileData[:])
}
func writeFileContentsString(path string, content string) {
	err := ioutil.WriteFile(path, []byte(content), 777)
	utils.HandleError(err)
}
