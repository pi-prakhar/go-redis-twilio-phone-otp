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

	//Parse and validate json body from request
	if err := utils.ParseAndValidateBody(r, &data); err != nil {
		utils.Log.Info("Error : Failed to parse json body from request")
		res = response.ErrorResponse{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusBadRequest)
		return
	}
	utils.Log.Info("Successfully Parsed and validated json body from request")

	//Handle locked phone number efficiently
	isLocked, ttl, err := GetOTPLock(data.PhoneNumber)
	if err != nil {
		utils.Log.Info("Error : Failed to fetch lock data from cache")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	// if is locked return forbidden response with expiry time left
	if isLocked {
		utils.Log.Info("Phone number is locked")
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
	utils.Log.Info("Phone number is not locked")

	//create OTP Message
	OTPCode := utils.CreateOTPString(6)
	utils.Log.Info("Successfully created OTP Code")

	//put otp in cache
	if err := SetOTPInCache(data.PhoneNumber, OTPCode); err != nil {
		utils.Log.Info("Error : Failed to store OTP in cache ")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	utils.Log.Info("Successfully stored OTP code in cache")

	//send otp to phone number : catch error
	if _, err := SendOTPMessage(data.PhoneNumber, OTPCode); err != nil {
		utils.Log.Info("Error : Failed to send OTP message")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	utils.Log.Info("Successfully send OTP message to user")

	//check number of tries in cache if empty set max tries
	otpTrials, err := GetOTPTrialsLeft(data.PhoneNumber)
	if err != nil {
		utils.Log.Info("Error : Failed to fetch OTP trials left from cache")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}

	//Check if OTP trials not set in cache
	if otpTrials == -1 {
		//set max otp trials
		otpTrials, err = SetMaxOTPTrials(data.PhoneNumber)
		if err != nil {
			utils.Log.Info("Error : Failed to set OTP trials left to max")
			res = response.ErrorResponse{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: string(err.Error()),
			}
			res.WriteJSON(w, http.StatusInternalServerError)
			return
		}
		utils.Log.Info("Successfully set OTP trials left to max")
	}
	utils.Log.Info(fmt.Sprintf("Successfully fetched OTP trials left : %d", otpTrials))
	//send otp send success message
	res = response.SuccessResponse[TrialsLeft]{
		StatusCode: http.StatusOK,
		Message:    "Successfully send OTP message",
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
	defer cancel()

	var data VerifyData
	var res response.Responder

	//bind json to otpdata model
	if err := utils.ParseAndValidateBody(r, &data); err != nil {
		utils.Log.Info("Error : Failed to parse json body from request")
		res = response.ErrorResponse{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusBadRequest)
		return
	}
	utils.Log.Info("Successfully Parsed and validated json body from request")

	//if locked
	isLocked, ttl, err := GetOTPLock(data.User.PhoneNumber)
	if err != nil {
		utils.Log.Info("Error : Failed to fetch lock data from cache")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	// if is locked return forbidden response with expiry time left
	if isLocked {
		utils.Log.Info("Phone number is locked")
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
	utils.Log.Info("Phone number is not locked")

	// Get cached otp
	cachedOTP, err := GetCachedOTPCode(data.User.PhoneNumber)
	if err != nil {
		utils.Log.Info("Error : Failed to fetch OTP code from cache")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	utils.Log.Info("Fetched OTP code from cache")

	//if otp data not present in cache , it has expired
	if cachedOTP == "" {
		utils.Log.Info("OTP expired")
		//fetch trials left
		trialsLeft, err := GetOTPTrialsLeft(data.User.PhoneNumber)
		if err != nil {
			utils.Log.Info("Error : Failed to fetch OTP trials left from cache")
			res = response.ErrorResponse{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: string(err.Error()),
			}
			res.WriteJSON(w, http.StatusInternalServerError)
			return
		}
		utils.Log.Info("Successfully fetched OTP trials left from cache")

		//if trial left is 1, Max limit reached
		if trialsLeft == 1 {
			utils.Log.Info("Max trial Limit reached")

			//perform cleanup
			if err := CleanUp(data.User.PhoneNumber); err != nil {
				utils.Log.Info("Error : Failed to set OTP lock in cache")
				res = response.ErrorResponse{
					StatusCode:   http.StatusInternalServerError,
					ErrorMessage: string(err.Error()),
				}
				res.WriteJSON(w, http.StatusInternalServerError)
				return
			}
			utils.Log.Info("Successfully CleanedUp")

			//set otp lock for 30 min
			if err := SetOTPLock(data.User.PhoneNumber, true); err != nil {
				utils.Log.Info("Error : Failed to set OTP lock in cache")
				res = response.ErrorResponse{
					StatusCode:   http.StatusInternalServerError,
					ErrorMessage: string(err.Error()),
				}
				res.WriteJSON(w, http.StatusInternalServerError)
				return
			}
			utils.Log.Info("Successfully locked phone number")

			//send forbidden response with expiry time
			res = response.SuccessResponse[TimeData]{
				StatusCode: http.StatusForbidden,
				Message:    "OTP Expired and Max Limit Reached, Try after 30 min",
				Data: TimeData{
					User: data.User,
					TTL:  30,
				},
			}
			res.WriteJSON(w, http.StatusInternalServerError)
			return
		}

		//decrement the trials left
		if err = DecrementOTPTrialsLeft(data.User.PhoneNumber); err != nil {
			utils.Log.Info("Error : Failed to Decrement OTP trials left in cache")
			res = response.ErrorResponse{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: string(err.Error()),
			}
			res.WriteJSON(w, http.StatusInternalServerError)
			return
		}
		utils.Log.Info("Successfully decremented OTP trials left in cache")

		//forbidden response with trials left
		res = response.SuccessResponse[TrialsLeft]{
			StatusCode: http.StatusUnauthorized,
			Message:    "OTP Expired, Try Again",
			Data: TrialsLeft{
				User:   data.User,
				Trials: trialsLeft - 1,
			},
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	//if otp != otp in cache
	if cachedOTP != data.Code {
		utils.Log.Info("Incorrect OTP")
		//get trials left
		trialsLeft, err := GetOTPTrialsLeft(data.User.PhoneNumber)
		if err != nil {
			utils.Log.Info("Error : Failed to fetch OTP trials left from cache")
			res = response.ErrorResponse{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: string(err.Error()),
			}
			res.WriteJSON(w, http.StatusInternalServerError)
			return
		}
		utils.Log.Info("Successfully fetched OTP trials left from cache")

		//if trials left is 1, max limit reached
		if trialsLeft == 1 {
			utils.Log.Info("Max trial limit reached")

			//perform cleanup
			if err := CleanUp(data.User.PhoneNumber); err != nil {
				utils.Log.Info("Error : Failed to set OTP lock in cache")
				res = response.ErrorResponse{
					StatusCode:   http.StatusInternalServerError,
					ErrorMessage: string(err.Error()),
				}
				res.WriteJSON(w, http.StatusInternalServerError)
				return
			}
			utils.Log.Info("Successfully CleanedUp")

			//set otp lock for 30 min
			if err := SetOTPLock(data.User.PhoneNumber, true); err != nil {
				utils.Log.Info("Error : Failed to set OTP lock in cache")
				res = response.ErrorResponse{
					StatusCode:   http.StatusInternalServerError,
					ErrorMessage: string(err.Error()),
				}
				res.WriteJSON(w, http.StatusInternalServerError)
				return
			}
			utils.Log.Info("Successfully locked phone number")

			//forbidden response with expiry time
			res = response.SuccessResponse[TimeData]{
				StatusCode: http.StatusForbidden,
				Message:    "Incorrect OTP and Max Limit Reached Try after 30 min",
				Data: TimeData{
					User: data.User,
					TTL:  30,
				},
			}
			res.WriteJSON(w, http.StatusInternalServerError)
			return
		}
		//decrement the trials left
		if err = DecrementOTPTrialsLeft(data.User.PhoneNumber); err != nil {
			utils.Log.Info("Error : Failed to Decrement OTP trials left in cache")
			res = response.ErrorResponse{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: string(err.Error()),
			}
			res.WriteJSON(w, http.StatusInternalServerError)
			return
		}
		utils.Log.Info("Successfully decremented OTP trials left in cache")
		//forbidden response with trials left
		res = response.SuccessResponse[TrialsLeft]{
			StatusCode: http.StatusUnauthorized,
			Message:    "Incorrect OTP, Try Again",
			Data: TrialsLeft{
				User:   data.User,
				Trials: trialsLeft - 1,
			},
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}

	utils.Log.Info("User is successfully verified")
	//perform clean up >delete cached otp > delete cached trials
	if err := CleanUp(data.User.PhoneNumber); err != nil {
		utils.Log.Info("Error : Failed to set OTP lock in cache")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	utils.Log.Info("Successfully CleanedUp")

	//send success message
	res = response.SuccessResponse[string]{
		StatusCode: http.StatusOK,
		Message:    "Successfully verified user",
		Data:       data.User.PhoneNumber,
	}
	res.WriteJSON(w, http.StatusOK)
}
