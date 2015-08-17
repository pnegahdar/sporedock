package utils

import (
	"fmt"
	logger "github.com/apsdehal/go-logger"
	"os"
	"runtime/debug"
)

var log, _ = logger.New("main", 1, os.Stdout)

func HandleError(err error) {
	if err != nil {
		log.Error(err.Error())
		debug.PrintStack()
		panic(err)
	}
}

func LogInfo(message string) {
	log.Info(message)
}

func LogInfoF(message string, a ...interface{}) {
	LogInfo(fmt.Sprintf(message, a...))
}

func LogWarn(message string) {
	log.Warning(message)
}

func LogWarnF(message string, a ...interface{}) {
	LogWarn(fmt.Sprintf(message, a...))
}

func LogDebug(message string) {
	log.Debug(message)
}
