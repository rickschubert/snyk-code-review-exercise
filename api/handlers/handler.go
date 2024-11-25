package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Handler interface {
	Handle() func(w http.ResponseWriter, r *http.Request)
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func SendErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	errResp, marshalErr := json.Marshal(ErrorResponse{
		Message: err.Error(),
	})
	if marshalErr != nil {
		println(fmt.Sprint("unable to marshal error into error response: %w", err))
	}

	w.WriteHeader(statusCode)
	w.Write(errResp)
}
