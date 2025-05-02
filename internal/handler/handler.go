package handler

import (
	"net/http"

	"github.com/torderonex/load-balancer/internal/storage"
)

// не стал выносить логику в сервисный слой, т.к. лоигики по сути и нет
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

	router.HandleFunc("/client/rate", h.SetClientRate)
	router.HandleFunc("/client/capacity", h.SetClientCapacity)

	return router
}
