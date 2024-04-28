package config

import (
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/utils"
	"github.com/pi-prakhar/utils/loader"
	"github.com/twilio/twilio-go"
)

func GetTwilioClient() *twilio.RestClient {
	username := getAccountSID()
	password := getAuthToken()

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: username,
		Password: password,
	})

	return client
}

func getAccountSID() string {
	account_sid, err := loader.GetValueFromEnv("TWILIO_ACCOUNT_SID")
	if err != nil {
		utils.Log.Error("Error fetching twilio account_sid", err)
	}
	return account_sid
}

func getAuthToken() string {
	auth_token, err := loader.GetValueFromEnv("TWILIO_AUTHTOKEN")
	if err != nil {
		utils.Log.Error("Error fetching twilio auth_token", err)
	}
	return auth_token
}

func GetTwilioPhoneNumber() string {
	phone_number, err := loader.GetValueFromEnv("TWILIO_PHONE_NUMBER")
	if err != nil {
		utils.Log.Error("Error fetching twilio phone_number", err)
	}
	return phone_number
}
