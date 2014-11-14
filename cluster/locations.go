package cluster

func GetAppLocationKey(appName string) string {
	return AppLocationsDirKey + appName + "/"
}

func GetMachineAppLocationKey(appName string, machineName string) string {
	return GetAppLocationKey(appName) + machineName
}
