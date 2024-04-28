package response

import "net/http"

type Responder interface {
	WriteJSON(http.ResponseWriter, int) error
}
