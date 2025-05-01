package strategy

import (
	"sync/atomic"

	"github.com/torderonex/load-balancer/internal/backend"
)

type roundRobinStrategy struct {
	current atomic.Int32
}

func NewRoundRobinStrategy() Strategy {
	return &roundRobinStrategy{
		current: atomic.Int32{},
	}
}

func (rr *roundRobinStrategy) NextBackend(backends []*backend.Backend) *backend.Backend {
	// атомарное инкрементирование счетчика
	// выбор следующего доступного бэкенда
	rr.current.Add(1)
	return backends[rr.current.Load()%int32(len(backends))]
}
