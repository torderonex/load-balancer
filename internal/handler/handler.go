package handler

import (
	"fmt"
	"log/slog"
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

func (h *Handler) InitRoutes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("/client/rate", h.SetClientRate)
	router.HandleFunc("/client/capacity", h.SetClientCapacity)
	router.HandleFunc("/client", h.GetClient)

	return h.logMiddleware(router)
}

func (h *Handler) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug(fmt.Sprintf("Request: %s %s", r.Method, r.URL.Path))
		next.ServeHTTP(w, r)
	})
}
