package response

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse[T any] struct {
	Status  int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"` // Optional data field
}

func (sr SuccessResponse[T]) WriteJSON(w http.ResponseWriter, code int) error {
	jsonData, err := json.Marshal(sr)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_, err = w.Write(jsonData)
	return err
}
