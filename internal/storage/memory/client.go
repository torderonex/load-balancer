package memory

import (
	"sync"

	"github.com/torderonex/load-balancer/internal/model"
)

// MemoryStorage реализация хранилища клиентов в памяти
type clientStorage struct {
	clients map[string]*model.Client
	mu      sync.RWMutex
}

func NewClientStorage() *clientStorage {
	return &clientStorage{
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

func (ms *clientStorage) SetClientRate(clientID string, rate int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.clients[clientID].Rate = rate
}

func (ms *clientStorage) SetClientCapacity(clientID string, capacity int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.clients[clientID].Capacity = capacity
}

func (ms *clientStorage) SetClientStatus(clientID string, isBanned bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.clients[clientID].IsBanned = isBanned
}
