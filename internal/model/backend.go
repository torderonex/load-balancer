package model

import (
	"net/http"
	"net/url"
	"sync"
)

type Backend struct {
	URL       *url.URL
	Alive     bool
	LastError error
	mu        sync.RWMutex
}

func NewBackend(url *url.URL) *Backend {
	return &Backend{
		URL:       url,
		Alive:     true,
		LastError: nil,
		mu:        sync.RWMutex{},
	}
}

func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Alive = alive
}

func (b *Backend) IsAlive() bool {
	return b.Alive
}

// Считаем сервер живым, если он отвечает любым статус-кодом
func (b *Backend) CheckHealth() bool {
	resp, err := http.Get(b.URL.String())
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return true
}
