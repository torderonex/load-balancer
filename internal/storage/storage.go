package storage

import (
	"github.com/torderonex/load-balancer/internal/model"
	"github.com/torderonex/load-balancer/internal/storage/memory"
)

type Storage struct {
	ClientStorage
}

func New() *Storage {
	return &Storage{
		ClientStorage: memory.NewClientStorage(),
	}
}

type ClientStorage interface {
	GetClient(clientID string) (*model.Client, bool)
	SaveClient(clientID string, client *model.Client)
	DeleteClient(clientID string) error
	GetAllClientIDs() ([]string, error)
	SetClientRate(clientID string, rate int)
	SetClientCapacity(clientID string, capacity int)
	SetClientStatus(clientID string, isBanned bool)
}
