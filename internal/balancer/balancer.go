package balancer

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/torderonex/load-balancer/internal/backend"
	"github.com/torderonex/load-balancer/internal/balancer/strategy"
	"github.com/torderonex/load-balancer/pkg/sl"
)

type Balancer struct {
	backends []*backend.Backend
	strategy strategy.Strategy
	proxy    *httputil.ReverseProxy
	mu       sync.RWMutex
}

func NewBalancer(strategyName string) *Balancer {
	return &Balancer{
		strategy: strategy.NewStrategy(strategyName),
	}
}

func (b *Balancer) AddBackends(urls []string) {
	for _, u := range urls {
		tmp, err := url.Parse(u)
		if err != nil {
			slog.Error("Ошибка при парсинге URL", sl.Err(err))
			continue
		}
		b.backends = append(b.backends, &backend.Backend{URL: tmp})
	}
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientID := r.RemoteAddr

	//TODO: Проверка rate limit

	slog.Info(fmt.Sprintf("Получен запрос: %s %s от %s", r.Method, r.URL.Path, clientID))

	// Получение бэкенда по стратегии
	backend := b.strategy.NextBackend(b.backends)
	if backend == nil {
		slog.Error("Нет доступных бэкендов")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	// Проксирование запроса
	proxy := httputil.NewSingleHostReverseProxy(backend.URL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("Ошибка проксирования к %s: %v", backend.URL.String(), err)
		backend.SetAlive(false)
		backend.LastError = err

		// Повторная попытка с другим бэкендом
		b.ServeHTTP(w, r)
	}

	proxy.ServeHTTP(w, r)
}

func (b *Balancer) StartHealthCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slog.Debug("Проверка доступности бэкендов")
			b.mu.RLock()
			for _, back := range b.backends {
				go func(b *backend.Backend) {
					alive := b.CheckHealth()
					if b.IsAlive() != alive {
						slog.Info(fmt.Sprintf("Изменение статуса бэкенда %s: %v -> %v",
							b.URL.String(), b.IsAlive(), alive))
					}
					b.SetAlive(alive)
				}(back)
			}
			b.mu.RUnlock()
		}
	}
}
