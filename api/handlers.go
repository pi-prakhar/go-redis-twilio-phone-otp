package api

import (
	"context"
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
		res = response.ErrorResponse{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusBadRequest)
		return
	}

	//TODO : create rate limit
	//create OTP Message
	OTPCode := utils.CreateOTPString(6)

	//put otp in cache
	if err := StoreOTPInCache(data.PhoneNumber, OTPCode); err != nil {
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
	otpTrials, err := OTPTrialsLeft(data.PhoneNumber)
	if err != nil {
		utils.Log.Info("Failed to fetch otp trials")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	if otpTrials == -1 {
		otpTrials, err = SetMaxOTPTrials(data.PhoneNumber)
	}
	if err != nil {
		utils.Log.Info("Failed to set max otp trials")
		res = response.ErrorResponse{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusInternalServerError)
		return
	}
	//send otp send success message

	res = response.SuccessResponse[int]{
		Status:  http.StatusOK,
		Message: "Successfully send otp message",
		Data:    otpTrials,
	}
	res.WriteJSON(w, http.StatusOK)
}

//handler funtion fot verify otp

func VerifyOTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	_, cancel := context.WithTimeout(context.Background(), appTimeout)
	// var payload data.VerifyData
	defer cancel()

	var data VerifyData
	var res response.Responder

	//bind json to otpdata model
	if err := utils.ParseAndValidateBody(r, &data); err != nil {
		res = response.ErrorResponse{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: string(err.Error()),
		}
		res.WriteJSON(w, http.StatusBadRequest)
		return
	}
	//if lock error 403
	//if otp data not present in cache
	//	//if trials left - 1 is zero -> error 403max limit reached data lock time(30 min) -> set lock
	//	//set trials -1
	//	// error message 403 otp expired trials left
	cachedOTP, err := GetCachedOTPCode(data.User.PhoneNumber)
	if err != nil {
		//TODO : give err response
		return
	}

	if cachedOTP == "" {
		trialsLeft, err := OTPTrialsLeft(data.User.PhoneNumber)
		if err != nil {
			//TODO : give err response
			return
		}
		if trialsLeft == 1 {

			//TODO : give err response
		}
	}
	//if otp == otp in cache
	//	//if trials left - 1 is zero -> error 403max limit reached data lock time(30 min) -> set lock
	//	//set trials -1
	//	// error message 403 otp expired trials left
	//delete cached otp
	//delete cached trials
	//send success message

}
