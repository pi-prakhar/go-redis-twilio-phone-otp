package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/internal/response"
	"github.com/pi-prakhar/go-redis-twilio-phone-otp/utils"
)

//router function with to api endpoints /send-otp /verify-otp

func New() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/hello-world", func(w http.ResponseWriter, r *http.Request) {
		res := response.SuccessResponse[string]{
			Status:  http.StatusOK,
			Message: "Successful Response",
			Data:    "Hello world",
		}
		res.WriteJSON(w, http.StatusOK)
		utils.Log.Info("Succesfull Response : Hello World")
	})
	r.HandleFunc("/api/send-otp", SendOTP)
	r.HandleFunc("/api/verify-otp", VerifyOTP)

	return r
}
