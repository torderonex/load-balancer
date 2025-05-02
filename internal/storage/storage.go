package storage

import (
	"github.com/torderonex/load-balancer/internal/config"
	"github.com/torderonex/load-balancer/internal/model"
	"github.com/torderonex/load-balancer/internal/storage/memory"
)

type Storage struct {
	ClientStorage
}

func New(cfg *config.Config) *Storage {
	return &Storage{
		ClientStorage: memory.NewClientStorage(&cfg.RateLimiter),
	}
}

type ClientStorage interface {
	GetClient(clientID string) (*model.Client, bool)
	SaveClient(clientID string, client *model.Client)
	DeleteClient(clientID string) error
	GetAllClientIDs() ([]string, error)
	SetClientRate(clientID string, rate int) error
	SetClientCapacity(clientID string, capacity int) error
}
