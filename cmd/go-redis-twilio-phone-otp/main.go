package main

import (
	"fmt"
	"net/http"
	"time"

	router "github.com/pi-prakhar/go-redis-twilio-phone-otp/api"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/utils"
	loader "github.com/pi-prakhar/utils/loader"
)

func init() {
	utils.InitLogger()
	utils.Log.Info("GO-PHONE-OTP-SERVICE Logger Started")

	err := loader.LoadEnv()

	if err != nil {
		utils.Log.Error("Failed to Load ENV", err)
	}
}

func main() {
	var domain string
	isProduction, err := loader.GetValueFromConf("production")

	if err != nil {
		utils.Log.Error("Failed to load config data", err)
	}

	if isProduction == "false" {
		port, err := loader.GetValueFromConf("test-port")

		if err != nil {
			utils.Log.Error("Failed to load config data", err)
		}

		domain = fmt.Sprintf(":%s", port)
	} else {
		domain, err = loader.GetValueFromConf("prod-domain")

		if err != nil {
			utils.Log.Error("Failed to load config data", err)
		}
	}

	srv := &http.Server{
		Handler:      router.New(),
		Addr:         domain,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	utils.Log.Error(fmt.Sprintf("could not start server at : %s", domain), srv.ListenAndServe())
}
