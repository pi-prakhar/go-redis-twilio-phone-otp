package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pi-prakhar/go-redis-twilio-phone-otp/internal/response"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/utils"
)

//handler function for send otp

const appTimeout = time.Second * 10

func SendOTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	_, cancel := context.WithTimeout(context.Background(), appTimeout)
	defer cancel()

	var data OTPData
	var res response.Responder
	//bind json to otpdata model
	if err := utils.ParseAndValidateBody(r, &data); err != nil {
		utils.Log.Debug("Error in Parsing Data")
		res = response.ErrorResponse{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusBadRequest)
		return
	}

	//Check if phone number is locked
	isLocked, ttl, err := GetOTPLock(data.PhoneNumber)
	if err != nil {
		utils.Log.Info("Failed to fetch lock data from cache")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	// if is locked return forbidden response with expiry time left
	if isLocked {
		utils.Log.Info("Current phone number is locked for")
		res = response.SuccessResponse[TimeData]{
			StatusCode: http.StatusForbidden,
			Message:    fmt.Sprintf("User is prohibted to make any OTP request, Try after %d minutes", ttl),
			Data: TimeData{
				User: &data,
				TTL:  ttl,
			},
		}
		res.WriteJSON(w, http.StatusForbidden)
		return
	}
	//create OTP Message
	OTPCode := utils.CreateOTPString(6)

	//put otp in cache
	if err := SetOTPInCache(data.PhoneNumber, OTPCode); err != nil {
		utils.Log.Info("Failed to store OTP in cache ")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	//send otp to phone number : catch error
	if _, err := SendOTPMessage(data.PhoneNumber, OTPCode); err != nil {
		utils.Log.Info("Failed to send OTP message")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}

	//check number of tries in cache if empty set max tries
	otpTrials, err := GetOTPTrialsLeft(data.PhoneNumber)
	if err != nil {
		utils.Log.Info("Failed to fetch otp trials")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	//check if otp trials not set
	if otpTrials == -1 {
		//set max otp trials
		utils.Log.Debug("Set OTP trials to max")
		otpTrials, err = SetMaxOTPTrials(data.PhoneNumber)
		if err != nil {
			utils.Log.Info("Failed to set max otp trials")
			res = response.ErrorResponse{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: string(err.Error()),
			}
			res.WriteJSON(w, http.StatusInternalServerError)
			return
		}
	}

	//send otp send success message
	res = response.SuccessResponse[TrialsLeft]{
		StatusCode: http.StatusOK,
		Message:    "Successfully send otp message",
		Data: TrialsLeft{
			User:   &data,
			Trials: otpTrials,
		},
	}
	res.WriteJSON(w, http.StatusOK)
}

// handler funtion fot verify otp
func VerifyOTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	_, cancel := context.WithTimeout(context.Background(), appTimeout)
	// var payload data.VerifyData
	defer cancel()

	var data VerifyData
	var res response.Responder

	//bind json to otpdata model
	if err := utils.ParseAndValidateBody(r, &data); err != nil {
		utils.Log.Debug("Error in Parsing Data")
		res = response.ErrorResponse{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusBadRequest)
		return
	}
	//if locked
	isLocked, ttl, err := GetOTPLock(data.User.PhoneNumber)
	if err != nil {
		utils.Log.Info("Failed to fetch lock data from cache")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	// if is locked return forbidden response with expiry time left
	if isLocked {
		utils.Log.Info("Current phone number is locked for")
		res = response.SuccessResponse[TimeData]{
			StatusCode: http.StatusForbidden,
			Message:    fmt.Sprintf("User is prohibted to make any OTP request, Try after %d minutes", ttl),
			Data: TimeData{
				User: data.User,
				TTL:  ttl,
			},
		}
		res.WriteJSON(w, http.StatusForbidden)
		return
	}
	// Get cached otp
	cachedOTP, err := GetCachedOTPCode(data.User.PhoneNumber)
	if err != nil {
		utils.Log.Info("Failed to fetch cached otp")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	//if otp data not present in cache , it has expired
	if cachedOTP == "" {
		utils.Log.Debug("OTP expired")
		//fetch trials left
		trialsLeft, err := GetOTPTrialsLeft(data.User.PhoneNumber)
		if err != nil {
			utils.Log.Info("Failed to fetch trials left")
			res = response.ErrorResponse{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: string(err.Error()),
			}
			res.WriteJSON(w, http.StatusInternalServerError)
			return
		}
		//if trial left is 1, Max limit reached
		if trialsLeft == 1 {
			utils.Log.Info("Max Limit reached")
			//set otp lock for 30 min
			if err := SetOTPLock(data.User.PhoneNumber, true); err != nil {
				utils.Log.Info("Failed to set otp lock")
				res = response.ErrorResponse{
					StatusCode:   http.StatusInternalServerError,
					ErrorMessage: string(err.Error()),
				}
				res.WriteJSON(w, http.StatusInternalServerError)
				return
			}
			//decrement trials left by 1
			if err := DecrementOTPTrials(data.User.PhoneNumber); err != nil {
				utils.Log.Info("Failed to decrement the value in cache")
				res = response.ErrorResponse{
					StatusCode:   http.StatusInternalServerError,
					ErrorMessage: string(err.Error()),
				}
				res.WriteJSON(w, http.StatusInternalServerError)
				return
			}
			//send forbidden response with expiry time
			res = response.SuccessResponse[TimeData]{
				StatusCode: http.StatusForbidden,
				Message:    "Max Limit Reached, Try after 30 min",
				Data: TimeData{
					User: data.User,
					TTL:  30,
				},
			}
			res.WriteJSON(w, http.StatusInternalServerError)
			return
		}
		//forbidden response with trials left
		res = response.SuccessResponse[TrialsLeft]{
			StatusCode: http.StatusForbidden,
			Message:    "OTP Expired, Try Again",
			Data: TrialsLeft{
				User:   data.User,
				Trials: trialsLeft,
			},
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	//if otp != otp in cache
	if cachedOTP != data.Code {
		//get trials left
		trialsLeft, err := GetOTPTrialsLeft(data.User.PhoneNumber)
		if err != nil {
			utils.Log.Info("Failed to fetch trials left")
			res = response.ErrorResponse{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: string(err.Error()),
			}
			res.WriteJSON(w, http.StatusInternalServerError)
			return
		}
		//if trials left is 1, max limit reached
		if trialsLeft == 1 {
			utils.Log.Info("Max Limit reached")
			//set lock
			if err := SetOTPLock(data.User.PhoneNumber, true); err != nil {
				utils.Log.Info("Failed to set otp lock")
				res = response.ErrorResponse{
					StatusCode:   http.StatusInternalServerError,
					ErrorMessage: string(err.Error()),
				}
				res.WriteJSON(w, http.StatusInternalServerError)
				return
			}
			//decrement trials by 1
			if err := DecrementOTPTrials(data.User.PhoneNumber); err != nil {
				utils.Log.Info("Failed to decrement the value in cache")
				res = response.ErrorResponse{
					StatusCode:   http.StatusInternalServerError,
					ErrorMessage: string(err.Error()),
				}
				res.WriteJSON(w, http.StatusInternalServerError)
				return
			}
			//forbidden response with expiry time
			res = response.SuccessResponse[TimeData]{
				StatusCode: http.StatusForbidden,
				Message:    "Max Limit Reached Try after 30 min",
				Data: TimeData{
					User: data.User,
					TTL:  30,
				},
			}
			res.WriteJSON(w, http.StatusInternalServerError)
			return
		}
		//forbidden response with trials left
		res = response.SuccessResponse[TrialsLeft]{
			StatusCode: http.StatusForbidden,
			Message:    "OTP Expired, Try Again",
			Data: TrialsLeft{
				User:   data.User,
				Trials: trialsLeft,
			},
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}

	//clean up
	//delete cached otp
	//delete cached trials
	CleanUp(data.User.PhoneNumber)

	//send success message
	res = response.SuccessResponse[string]{
		StatusCode: http.StatusOK,
		Message:    "Successfully verified OTP",
		Data:       data.User.PhoneNumber,
	}
	res.WriteJSON(w, http.StatusOK)

}
