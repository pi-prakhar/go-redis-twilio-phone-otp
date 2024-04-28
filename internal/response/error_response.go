package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	StatusCode   int    `json:"code"`
	ErrorMessage string `json:"message"`
}

func (er ErrorResponse) WriteJSON(w http.ResponseWriter, code int) error {
	jsonData, err := json.Marshal(er)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_, err = w.Write(jsonData)
	return err
}
