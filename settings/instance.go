package settings

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/pnegahdar/sporedock/utils"
	"io/ioutil"
)

var INSTANCE_NAME = get_instance_name()


func get_instance_name() string {
	filePath := GetInstanceIdConfPath()
	fileData, err := ioutil.ReadFile(filePath)
	utils.HandleError(err)
	if fileData == nil{
		uuidBase := uuid.NewRandom()
		uuidString := uuid.NewSHA1(uuidBase, nil).String()
		ioutil.WriteFile(filePath, byte(uuidString), 0644)
	}
	return string(fileData)
}
func GetDiscoveryString() string{
 return ""
}
func SetDiscoveryString(disocvery string){
}
