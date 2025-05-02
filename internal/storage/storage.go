package storage

import (
	"github.com/torderonex/load-balancer/internal/config"
	"github.com/torderonex/load-balancer/internal/model"
	"github.com/torderonex/load-balancer/internal/storage/memory"
	"github.com/torderonex/load-balancer/internal/storage/redis"
)

type Storage struct {
	ClientStorage
}

func New(cfg *config.Config) *Storage {
	switch cfg.Storage.Type {
	case "redis":
		return &Storage{
			ClientStorage: redis.NewClientStorage(&cfg.RateLimiter, &cfg.Storage.Redis),
		}
	}
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
	Close() error
}
