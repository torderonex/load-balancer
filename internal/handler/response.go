package handler

import (
	"encoding/json"
	"errors"
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

// request errors
var (
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrInvalidAction    = errors.New("invalid action. Use 'ban' or 'unban'")
	ErrClientIDRequired = errors.New("client ID is required")
	ErrCapacityRequired = errors.New("capacity must be greater than 0")
	ErrClientNotFound   = errors.New("client not found")
)

// json parse errors
var (
	ErrInvalidContentType = errors.New("invalid content type")
	ErrReadRequestBody    = errors.New("error reading request body")
	ErrInvalidJSONFormat  = errors.New("invalid JSON format")
)
