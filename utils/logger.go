package utils

import (
	"fmt"

	"github.com/pi-prakhar/utils/loader"
	loggerUtil "github.com/pi-prakhar/utils/logger"
)

var (
	Log loggerUtil.Logger
)

func InitLogger() {
	serviceName, err := loader.GetValueFromConf("service_name")
	Log = loggerUtil.New(loggerUtil.DEBUG, serviceName)
	if err != nil {
		Log.Warn(fmt.Sprintf("%s", err))
	}
}
