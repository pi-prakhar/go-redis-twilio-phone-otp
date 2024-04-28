package api

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/internal/config"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/internal/database"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/utils"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

const OTP_CODE = "otp_code"
const OTP_LOCK = "lock"
const OTP_TRIAL_LEFT = "otp_trial_left"

// send otp message
func SendOTPMessage(phoneNumber string, OTPCode string) (string, error) {
	// _, cancel := context.WithTimeout(context.Background(), appTimeout)
	// defer cancel()
	twilioClient := config.GetTwilioClient()
	twilioPhoneNumber := config.GetTwilioPhoneNumber()
	messageString := fmt.Sprintf("OTP message is %s", OTPCode)
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(phoneNumber)
	params.SetFrom(twilioPhoneNumber)
	params.SetBody(messageString)

	res, err := twilioClient.Api.CreateMessage(params)
	if err != nil {
		utils.Log.Debug("Failed to send OTP message to user")
		return "", err
	}

	return *res.Sid, nil
}

// validate otp message
func ValidateOTPMessage(phoneNumber string, OTPCode string) (bool, error) {
	key := fmt.Sprintf("%s_%s", phoneNumber, OTP_CODE)
	cachedOTP, err := getCachedData(key)
	if err != nil {
		utils.Log.Debug("Failed to get value from cache")
		return false, err
	}

	if cachedOTP == nil {
		utils.Log.Debug("Key not present in cache")
		return false, err
	}

	if cachedOTP != OTPCode {
		utils.Log.Debug("OTP do not match")
		return false, nil
	}
	utils.Log.Info("OTP Verified")
	return true, nil
}

// store otp in cache
func StoreOTPInCache(phoneNumber string, OTPCode string) error {
	key := fmt.Sprintf("%s_%s", phoneNumber, OTP_CODE)
	otpTimeout, err := utils.GetOTPTimeout()
	if err != nil {
		utils.Log.Debug("Failed to fetch otp_timeout")
		return err
	}

	err = storeInCache(key, OTPCode, otpTimeout)
	if err != nil {
		utils.Log.Debug("Failed to store otp in cache")
		return err
	}
	utils.Log.Info("Successfully stored otp in cache")
	return nil
}

// get otp max tries left

func OTPTrialsLeft(phoneNumber string) (int, error) {
	key := fmt.Sprintf("%s_%s", phoneNumber, OTP_TRIAL_LEFT)
	otpTrialsLeft, err := getCachedData(key)

	if err != nil {
		utils.Log.Debug("Failed to fetch otp trials left from cache")
		return -1, err
	}

	if otpTrialsLeft == nil {
		otp_timeout, err := utils.GetOTPTimeout()
		if err != nil {
			utils.Log.Debug("Failed to fetch otp trials left from cache")
			return -1, err
		}
		otpMaxTrials, err := utils.GetOTPMaxTrials()
		if err != nil {
			utils.Log.Debug("Failed to fetch otp trials left from cache")
			return -1, err
		}
		err = storeInCache(key, otpMaxTrials, otp_timeout)
		if err != nil {
			utils.Log.Debug("Failed to set max trials in cache")
			return -1, err
		}
		return otpMaxTrials, nil
	}
	otpTrialsLeftInt, _ := otpTrialsLeft.(int)
	return otpTrialsLeftInt, nil

}

// check of rate limit

// store otp in cache
func storeInCache(key string, value any, expiry time.Duration) error {
	rdb := database.Client(0)
	ctx := database.Ctx

	err := rdb.Set(ctx, key, value, expiry).Err()
	if err != nil {
		utils.Log.Debug("Failed to store data in cache")
		return err
	}
	utils.Log.Info("Successfully Stored data in cache")
	return nil
}

// get data from cache
func getCachedData(key string) (any, error) {
	rdb := database.Client(0)
	ctx := database.Ctx
	cachedData, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		utils.Log.Debug("Key not present in cache")
		return nil, nil
	} else if err != nil {
		utils.Log.Debug("Failed to get value from cache")
		return nil, err
	} else {
		utils.Log.Info("Key is present in cache")
		return cachedData, nil
	}
}

// decrement value of a key from cache
func decrementValueInCache(key string) error {
	rdb := database.Client(0)
	ctx := database.Ctx
	err := rdb.Decr(ctx, key).Err()
	if err != nil {
		utils.Log.Debug("Failed to decrement value from cache")
		return err
	} else {
		utils.Log.Info("Successfully decremented value from cache")
		return nil
	}
}

// increment value of a key from cache
func incrementValueInCache(key string) error {
	rdb := database.Client(0)
	ctx := database.Ctx
	err := rdb.Incr(ctx, key).Err()
	if err != nil {
		utils.Log.Debug("Failed to decrement value from cache")
		return err
	} else {
		utils.Log.Info("Successfully decremented value from cache")
		return nil
	}
}
