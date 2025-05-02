package balancer

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/torderonex/load-balancer/internal/balancer/strategy"
	"github.com/torderonex/load-balancer/internal/limiter"
	"github.com/torderonex/load-balancer/internal/model"
	"github.com/torderonex/load-balancer/pkg/sl"
)

type Balancer struct {
	limiter  limiter.Limiter
	backends []*model.Backend
	strategy strategy.Strategy
	proxy    *httputil.ReverseProxy
	mu       sync.RWMutex
}

func NewBalancer(strategyName string, limiter limiter.Limiter) *Balancer {
	return &Balancer{
		limiter:  limiter,
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
		b.backends = append(b.backends, model.NewBackend(tmp))
	}
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientAddr := r.RemoteAddr
	clientIP := clientAddr

	// Обработка IPv4: "127.0.0.1:1234" -> "127.0.0.1"
	if idx := strings.LastIndex(clientAddr, ":"); idx != -1 && !strings.HasPrefix(clientAddr, "[") {
		clientIP = clientAddr[:idx]
	}

	// Обработка IPv6: "[::1]:1234" -> "::1"
	if strings.HasPrefix(clientAddr, "[") && strings.Contains(clientAddr, "]:") {
		clientIP = strings.TrimPrefix(clientAddr, "[")
		if idx := strings.Index(clientIP, "]:"); idx != -1 {
			clientIP = clientIP[:idx]
		}
	}
	if !b.limiter.Allow(clientIP) {
		slog.Info(fmt.Sprintf("Клиент %s превысил лимит", clientIP))
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	slog.Info(fmt.Sprintf("Получен запрос: %s %s от %s", r.Method, r.URL.Path, clientAddr))

	backend := b.strategy.NextBackend(b.backends)
	if backend == nil {
		slog.Error("Нет доступных бэкендов")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	// Проксирование запроса
	slog.Info(fmt.Sprintf("Проксирование запроса к %s", backend.URL.String()), slog.String("from", clientIP))
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
				go func(b *model.Backend) {
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
