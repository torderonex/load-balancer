package limiter

import (
	"log/slog"
	"sync"
	"time"

	"github.com/torderonex/load-balancer/internal/config"
	"github.com/torderonex/load-balancer/internal/model"
	"github.com/torderonex/load-balancer/internal/storage"
)

// Client представляет собой отдельного клиента с собственным bucket токенов

type Limiter interface {
	Allow(clientID string) bool
	StartRefill(interval time.Duration)
}

// TokenBucket основной лимитер, использующий алгоритм Token Bucket
type TokenBucket struct {
	storage     storage.ClientStorage
	defaultRate int // токенов в секунду
	capacity    int // максимальное количество токенов
	mu          sync.RWMutex
}

// NewTokenBucket создает новый экземпляр TokenBucket
func NewTokenBucket(config *config.RateLimiter, storage *storage.Storage) Limiter {
	return &TokenBucket{
		storage:     storage.ClientStorage,
		defaultRate: config.DefaultRate,
		capacity:    config.Capacity,
	}
}

// Allow проверяет, разрешен ли запрос от данного клиента
// Возвращает true, если запрос разрешен, и false, если превышен лимит
func (tb *TokenBucket) Allow(clientID string) bool {
	// Получаем или создаем клиента
	client, exists := tb.storage.GetClient(clientID)
	slog.Debug("Проверка клиента", "clientID", clientID, "exists", exists)
	if !exists {
		// Новый клиент получает полный bucket
		client = &model.Client{
			Tokens:     tb.capacity,
			Capacity:   tb.capacity,
			Rate:       tb.defaultRate,
			LastRefill: time.Now(),
		}
		tb.storage.SaveClient(clientID, client)
		return true
	}
	slog.Info("client", "client", client)
	if client.IsBanned {
		return false
	}

	// Пополняем токены согласно прошедшему времени
	now := time.Now()
	elapsed := now.Sub(client.LastRefill).Seconds()
	client.LastRefill = now

	// Рассчитываем сколько токенов нужно добавить
	tokensToAdd := int(elapsed * float64(client.Rate))
	if tokensToAdd > 0 {
		client.Tokens = min(client.Tokens+tokensToAdd, client.Capacity)
	}

	// Проверяем и забираем токен
	if client.Tokens > 0 {
		client.Tokens--
		tb.storage.SaveClient(clientID, client)
		return true
	}

	return false
}

// StartRefill запускает периодическое пополнение токенов для всех клиентов
func (tb *TokenBucket) StartRefill(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			clientIDs, err := tb.storage.GetAllClientIDs()
			if err != nil {
				continue
			}

			now := time.Now()
			for _, id := range clientIDs {
				client, exists := tb.storage.GetClient(id)
				if !exists {
					continue
				}

				elapsed := now.Sub(client.LastRefill).Seconds()
				tokensToAdd := int(elapsed * float64(client.Rate))

				if tokensToAdd > 0 {
					client.Tokens = min(client.Tokens+tokensToAdd, client.Capacity)
					client.LastRefill = now
					tb.storage.SaveClient(id, client)
				}
			}
		}
	}()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
