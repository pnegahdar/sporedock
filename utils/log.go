package utils

import (
	logger "github.com/apsdehal/go-logger"
	"os"
)

var log, _ = logger.New("test", 1, os.Stdout)

func HandleError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func LogInfo(message string) {
	log.Info(message)
}

func LogWarn(message string) {
	log.Warning(message)
}

func LogDebug(message string) {
	log.Debug(message)
}
