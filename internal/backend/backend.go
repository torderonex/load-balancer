package backend

import (
	"net/url"
	"sync"
)

type Backend struct {
	URL       *url.URL
	Alive     bool
	LastError error
	mu        sync.RWMutex
}

func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Alive = alive
}

func (b *Backend) IsAlive() bool {
	return b.Alive
}

func (b *Backend) CheckHealth() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	alive := b.Alive
	b.Alive = true
	return alive
}
