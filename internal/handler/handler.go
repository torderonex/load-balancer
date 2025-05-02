package handler

import (
	"net/http"

	"github.com/torderonex/load-balancer/internal/storage"
)

type Handler struct {
	storage *storage.Storage
}

func NewHandler(storage *storage.Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

func (h *Handler) InitRoutes() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/client/status", h.SetClientStatus)
	router.HandleFunc("/client/rate", h.SetClientRate)
	router.HandleFunc("/client/capacity", h.SetClientCapacity)

	return router
}
