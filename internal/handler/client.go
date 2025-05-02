package handler

import (
	"encoding/json"
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

func (h *Handler) SetClientRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		newErrorResponse(w, r, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var req RateRequest
	if err := parseJSONBody(r, &req); err != nil {
		newErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if req.ClientID == "" {
		newErrorResponse(w, r, http.StatusBadRequest, ErrClientIDRequired)
		return
	}

	h.storage.SetClientRate(req.ClientID, req.Rate)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) SetClientCapacity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		newErrorResponse(w, r, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	var req CapacityRequest
	if err := parseJSONBody(r, &req); err != nil {
		newErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if req.ClientID == "" {
		newErrorResponse(w, r, http.StatusBadRequest, ErrClientIDRequired)
		return
	}

	if req.Capacity < 0 {
		newErrorResponse(w, r, http.StatusBadRequest, ErrCapacityRequired)
		return
	}

	h.storage.SetClientCapacity(req.ClientID, req.Capacity)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) GetClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		newErrorResponse(w, r, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
		return
	}

	clientID := r.URL.Query().Get("clientId")
	if clientID == "" {
		newErrorResponse(w, r, http.StatusBadRequest, ErrClientIDRequired)
		return
	}

	client, ok := h.storage.GetClient(clientID)
	if !ok {
		newErrorResponse(w, r, http.StatusNotFound, ErrClientNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}

// parseJSONBody вспомогательная функция для парсинга JSON тела запроса
func parseJSONBody(r *http.Request, v interface{}) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return ErrInvalidContentType
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		return ErrReadRequestBody
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, v); err != nil {
		return ErrInvalidJSONFormat
	}

	return nil
}
