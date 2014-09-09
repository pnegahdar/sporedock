package settings

import (
	"code.google.com/p/go-uuid/uuid"
	"io/ioutil"
)

func GetInstanceName() string {
	filePath := GetInstanceIdConfPath()
	data := getFileContentsString(filePath)
	if data == "" {
		uuidBase := uuid.NewRandom()
		uuidString := uuid.NewSHA1(uuidBase, nil).String()
		writeFileContentsString(filePath, uuidString)
		data = uuidString
	}
	return data
}
func GetDiscoveryString() string {
	filePath := GetDiscoveryConfPath()
	return getFileContentsString(filePath)
}
func SetDiscoveryString(discovery string) {
	filePath := GetDiscoveryConfPath()
	writeFileContentsString(filePath, discovery)
}

func getFileContentsString(path string) string {
	fileData, err := ioutil.ReadFile(path)
	if err {
		return ""
	}
	return string(fileData[:])
}
func writeFileContentsString(path string, content string) {
	ioutil.WriteFile(path, []byte(content), 0644)
}
