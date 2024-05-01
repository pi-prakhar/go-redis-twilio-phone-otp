package utils

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/pi-prakhar/utils/loader"
)

var validate = validator.New()

//func to verify if phone number is proper or not

// validate body
func ParseAndValidateBody(r *http.Request, data any) error {

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		Log.Debug("ERROR : Invalid request body")
		return err
	}

	if err := validate.Struct(data); err != nil {
		Log.Debug("Error : Failed to validate request json")
		return err
	}

	return nil
}

func GetOTPTimeout() (time.Duration, error) {
	otpTimeoutFromConf, err := loader.GetValueFromConf("otp-timeout")
	if err != nil {
		Log.Debug("Failed to load value from conf")
		return -1, err
	}
	otpTimeoutValInt, err := strconv.Atoi(otpTimeoutFromConf)
	if err != nil {
		Log.Debug("Failed to convert string to int")
		return -1, err
	}
	otpTimeout := time.Second * time.Duration(otpTimeoutValInt)
	return otpTimeout, nil

}

func GetLockTimeout() (time.Duration, error) {
	lockTimeoutFromConf, err := loader.GetValueFromConf("otp-lock-timeout")
	if err != nil {
		Log.Debug("Failed to load value from conf")
		return -1, err
	}
	lockTimeoutValInt, err := strconv.Atoi(lockTimeoutFromConf)
	if err != nil {
		Log.Debug("Failed to convert string to int")
		return -1, err
	}
	lockTimeout := time.Minute * time.Duration(lockTimeoutValInt)
	return lockTimeout, nil

}

func GetOTPMaxTrials() (int, error) {
	otpMaxTrialsFromConf, err := loader.GetValueFromConf("otp-max-trials")
	if err != nil {
		Log.Debug("Failed to load value from conf")
		return -1, err
	}
	otpMaxTrialsValInt, err := strconv.Atoi(otpMaxTrialsFromConf)
	if err != nil {
		Log.Debug("Failed to convert string to int")
		return -1, err
	}

	return otpMaxTrialsValInt, nil

}

func CreateOTPString(maxDigits int) string {
	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, maxDigits)
	n, err := io.ReadAtLeast(rand.Reader, b, maxDigits)
	if n != maxDigits {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}
