package memory

import (
	"sync"
	"time"

	"github.com/torderonex/load-balancer/internal/config"
	"github.com/torderonex/load-balancer/internal/model"
)

// MemoryStorage реализация хранилища клиентов в памяти
type clientStorage struct {
	clients map[string]*model.Client
	mu      sync.RWMutex
	cfg     *config.RateLimiter
}

func NewClientStorage(cfg *config.RateLimiter) *clientStorage {
	return &clientStorage{
		cfg:     cfg,
		clients: make(map[string]*model.Client),
	}
}

func (ms *clientStorage) GetAllClientIDs() ([]string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ids := make([]string, 0, len(ms.clients))
	for id := range ms.clients {
		ids = append(ids, id)
	}

	return ids, nil
}

func (ms *clientStorage) GetClient(clientID string) (*model.Client, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	client, exists := ms.clients[clientID]
	return client, exists
}

func (ms *clientStorage) SaveClient(clientID string, client *model.Client) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.clients[clientID] = client
}

func (ms *clientStorage) DeleteClient(clientID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.clients, clientID)
	return nil
}

func (ms *clientStorage) SetClientRate(clientID string, rate int) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, ok := ms.clients[clientID]; !ok {
		ms.clients[clientID] = &model.Client{
			Rate:       rate,
			Capacity:   ms.cfg.Capacity,
			LastRefill: time.Now(),
			IP:         clientID,
			Tokens:     ms.cfg.Capacity,
		}
	}

	ms.clients[clientID].Rate = rate
	return nil
}

func (ms *clientStorage) SetClientCapacity(clientID string, capacity int) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, ok := ms.clients[clientID]; !ok {
		ms.clients[clientID] = &model.Client{
			Rate:       ms.cfg.DefaultRate,
			Capacity:   capacity,
			LastRefill: time.Now(),
			IP:         clientID,
			Tokens:     ms.cfg.Capacity,
		}
	}

	ms.clients[clientID].Capacity = capacity
	ms.clients[clientID].Tokens = capacity
	return nil
}
