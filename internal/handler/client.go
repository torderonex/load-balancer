package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type StatusRequest struct {
	ClientID string `json:"clientId"`
	Action   string `json:"action"`
}

// RateRequest структура для парсинга запроса изменения скорости
type RateRequest struct {
	ClientID string `json:"clientId"`
	Rate     int    `json:"rate"`
}

// CapacityRequest структура для парсинга запроса изменения емкости
type CapacityRequest struct {
	ClientID string `json:"clientId"`
	Capacity int    `json:"capacity"`
}

func (h *Handler) SetClientStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		newErrorResponse(w, r, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	var req StatusRequest
	if err := parseJSONBody(r, &req); err != nil {
		newErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if req.ClientID == "" {
		newErrorResponse(w, r, http.StatusBadRequest, errors.New("client ID is required"))
		return
	}

	switch req.Action {
	case "ban":
		h.storage.SetClientStatus(req.ClientID, true)
	case "unban":
		h.storage.SetClientStatus(req.ClientID, false)
	default:
		newErrorResponse(w, r, http.StatusBadRequest, errors.New("invalid action. Use 'ban' or 'unban'"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) SetClientRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		newErrorResponse(w, r, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	var req RateRequest
	if err := parseJSONBody(r, &req); err != nil {
		newErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if req.ClientID == "" {
		newErrorResponse(w, r, http.StatusBadRequest, errors.New("client ID is required"))
		return
	}

	h.storage.SetClientRate(req.ClientID, req.Rate)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) SetClientCapacity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		newErrorResponse(w, r, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	var req CapacityRequest
	if err := parseJSONBody(r, &req); err != nil {
		newErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if req.ClientID == "" {
		newErrorResponse(w, r, http.StatusBadRequest, errors.New("client ID is required"))
		return
	}

	if req.Capacity <= 0 {
		newErrorResponse(w, r, http.StatusBadRequest, errors.New("capacity must be greater than 0"))
		return
	}

	h.storage.SetClientCapacity(req.ClientID, req.Capacity)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// parseJSONBody вспомогательная функция для парсинга JSON тела запроса
func parseJSONBody(r *http.Request, v interface{}) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return errors.New("Content-Type must be application/json")
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		return errors.New("error reading request body")
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, v); err != nil {
		return errors.New("invalid JSON format")
	}

	return nil
}
