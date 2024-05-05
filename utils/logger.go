package utils

import (
	"fmt"

	"github.com/pi-prakhar/utils/loader"
	loggerUtil "github.com/pi-prakhar/utils/logger"
)

var (
	Log loggerUtil.Logger
)

func getLogLevel() loggerUtil.LogLevel {
	logLevel, err := loader.GetValueFromConf("log-level")
	if err != nil {
		return loggerUtil.DEBUG
	}

	if logLevel == "debug" {
		return loggerUtil.DEBUG
	} else if logLevel == "info" {
		return loggerUtil.INFO
	} else if logLevel == "warn" {
		return loggerUtil.WARN
	} else {
		return loggerUtil.INFO
	}
}
func InitLogger() {
	serviceName, err := loader.GetValueFromConf("service_name")
	Log = loggerUtil.New(getLogLevel(), serviceName)
	if err != nil {
		Log.Warn(fmt.Sprintf("%s", err))
	}
}
