package utils

import (
	logger "github.com/apsdehal/go-logger"
	"os"
	"runtime/debug"
)

var log, _ = logger.New("main", 1, os.Stdout)

func HandleError(err error) {
	if err != nil {
		log.Error(err.Error())
		debug.PrintStack()
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
