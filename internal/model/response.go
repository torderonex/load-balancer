package model

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NewErrorResponse(w http.ResponseWriter, r *http.Request, status int, message error) {
	response := ErrorResponse{
		Message: message.Error(),
		Code:    status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
