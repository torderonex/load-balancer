package handler

import (
	"net/http"
	"strconv"
)

func (h *Handler) SetClientStatus(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	clientID := r.URL.Query().Get("clientID")

	if clientID == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
		return
	}

	switch action {
	case "ban":
		h.storage.SetClientStatus(clientID, true)
	case "unban":
		h.storage.SetClientStatus(clientID, false)
	default:
		http.Error(w, "Invalid action. Use 'ban' or 'unban'", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) SetClientRate(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("clientID")
	rate := r.URL.Query().Get("rate")

	if clientID == "" || rate == "" {
		http.Error(w, "Client ID and rate are required", http.StatusBadRequest)
		return
	}

	rateInt, err := strconv.Atoi(rate)
	if err != nil {
		http.Error(w, "Invalid rate", http.StatusBadRequest)
		return
	}

	h.storage.SetClientRate(clientID, rateInt)
}

func (h *Handler) SetClientCapacity(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("clientID")
	capacity := r.URL.Query().Get("capacity")

	if clientID == "" || capacity == "" {
		http.Error(w, "Client ID and capacity are required", http.StatusBadRequest)
		return
	}

	capacityInt, err := strconv.Atoi(capacity)
	if err != nil {
		http.Error(w, "Invalid capacity", http.StatusBadRequest)
		return
	}

	h.storage.SetClientCapacity(clientID, capacityInt)
}
