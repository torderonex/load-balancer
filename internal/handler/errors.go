package handler

import (
	"errors"
	"net/http"

	"github.com/torderonex/load-balancer/internal/model"
)

// чтобы не писать model.NewErrorResponse(w, r, status, message) каждый раз
func newErrorResponse(w http.ResponseWriter, r *http.Request, status int, message error) {
	model.NewErrorResponse(w, r, status, message)
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
