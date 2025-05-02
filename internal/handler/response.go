package handler

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func newErrorResponse(w http.ResponseWriter, r *http.Request, status int, message error) {
	response := errorResponse{
		Message: message.Error(),
		Code:    status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
