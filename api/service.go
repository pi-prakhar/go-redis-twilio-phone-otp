package api

import (
	"fmt"
	"strconv"

	"github.com/pi-prakhar/go-redis-twilio-phone-otp/internal/config"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/utils"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

func SendOTPMessage(phoneNumber string, OTPCode string) (string, error) {
	twilioClient := config.GetTwilioClient()
	twilioPhoneNumber := config.GetTwilioPhoneNumber()
	messageString := fmt.Sprintf("OTP message is %s", OTPCode)
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(phoneNumber)
	params.SetFrom(twilioPhoneNumber)
	params.SetBody(messageString)

	res, err := twilioClient.Api.CreateMessage(params)
	if err != nil {
		utils.Log.Debug("Error : Failed to send OTP message to user")
		return "", err
	}
	utils.Log.Debug("Successfully send OTP message to user")

	return *res.Sid, nil
}

func SetOTPInCache(phoneNumber string, OTPCode string) error {
	key := utils.GetOTPCodeKey(phoneNumber)
	otpTimeout, err := utils.GetOTPTimeout()
	if err != nil {
		utils.Log.Debug("Error : Failed to fetch OTP timeout from conf")
		return err
	}
	err = storeInCache(key, OTPCode, otpTimeout)
	if err != nil {
		utils.Log.Debug("Error : Failed to store OTP code in cache")
		return err
	}
	utils.Log.Info("Successfully stored OTP code in cache")
	return nil
}

func GetOTPTrialsLeft(phoneNumber string) (int, error) {
	key := utils.GetOTPTrialsLeftKey(phoneNumber)
	otpTrialsLeft, err := getCachedData(key)

	if err != nil {
		utils.Log.Debug("Error : Failed to fetch OTP trials left from cache")
		return -1, err
	}

	if otpTrialsLeft == nil {
		utils.Log.Debug("OTP trials left data not present in cache")
		return -1, nil
	}
	otpTrialsLeftString := otpTrialsLeft.(string)
	otpTrialsLeftInt, _ := strconv.Atoi(otpTrialsLeftString)
	utils.Log.Debug("Successfully fetched OTP trials left from cache")
	return otpTrialsLeftInt, nil
}

func SetMaxOTPTrials(phoneNumber string) (int, error) {
	key := utils.GetOTPTrialsLeftKey(phoneNumber)
	otpMaxTrials, err := utils.GetOTPMaxTrials()
	if err != nil {
		utils.Log.Debug("Error : Failed to load otp-max-trials from conf")
		return -1, err
	}
	err = storeInCacheNoExpiry(key, otpMaxTrials)
	if err != nil {
		utils.Log.Debug("Error : Failed to set OTP trials left to max in cache")
		return -1, err
	}

	utils.Log.Debug("Successfully set OTP trials left to max in cache")
	return otpMaxTrials, nil
}

func GetCachedOTPCode(phoneNumber string) (string, error) {
	key := utils.GetOTPCodeKey(phoneNumber)
	otpCode, err := getCachedData(key)

	if err != nil {
		utils.Log.Debug("Error : Failed to fetch OTP code from cache")
		return "", err
	}
	if otpCode == nil {
		utils.Log.Debug("OTP code not present in cache")
		return "", nil
	}
	otpCodeString, _ := otpCode.(string)

	utils.Log.Debug("Successfully fetched OTP code from cache")
	return otpCodeString, nil
}

func SetOTPLock(phoneNumber string, value bool) error {
	key := utils.GetOTPLockKey(phoneNumber)
	timeout, err := utils.GetLockTimeout()

	if err != nil {
		utils.Log.Debug("Error : Failed to fetch OTP lock timeout")
		return err
	}
	if err := storeInCache(key, value, timeout); err != nil {
		utils.Log.Debug("Error : Failed to store OTP lock in cache")
		return err
	}

	utils.Log.Debug("Successfully stored OTP lock in cache")
	return nil
}

func GetOTPLock(phoneNumber string) (bool, int, error) {
	key := utils.GetOTPLockKey(phoneNumber)
	lockValue, err := getCachedData(key)

	if err != nil {
		utils.Log.Debug("Error : Failed to get OTP lock data from cache")
		return false, -2, err
	}
	if lockValue == nil {
		utils.Log.Debug("Lock data not present in cache")
		return false, -2, nil
	}

	ttl, err := getTTLData(key)
	if err != nil {
		utils.Log.Debug("Error : Failed to load OTP lock ttl from cache")
		return false, -2, err
	}

	ttlInt := int(ttl.Minutes())
	loackValueBool := lockValue == "1"

	utils.Log.Debug("Successfully fetched OTP lock from cache")
	return loackValueBool, ttlInt, nil
}

func CleanUp(phoneNumber string) error {
	otpCodeKey := utils.GetOTPCodeKey(phoneNumber)
	otpTrialsLeftKey := utils.GetOTPTrialsLeftKey(phoneNumber)

	if err := deleteDataFromCache(otpCodeKey); err != nil {
		utils.Log.Debug("Error : Failed to delete OTP code from cache")
		return err
	}
	utils.Log.Debug("Successfully deleted OTP code from cache")

	if err := deleteDataFromCache(otpTrialsLeftKey); err != nil {
		utils.Log.Debug("Error : Failed to delete OTP Trials left from cache")
		return err
	}
	utils.Log.Debug("Successfully deleted OTP trials left from cache")
	return nil
}

func DecrementOTPTrialsLeft(phoneNumber string) error {
	key := utils.GetOTPTrialsLeftKey(phoneNumber)
	err := decrementValueInCache(key)

	if err != nil {
		utils.Log.Debug("Error : Failed to decrement OTP trials left in cache")
		return err
	}
	utils.Log.Debug("Successfully decremented OTP trials left in cache")
	return nil
}
