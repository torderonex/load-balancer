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
	if len(backends) == 0 {
		return nil
	}

	// атомарное инкрементирование счетчика
	// выбор следующего доступного бэкенда
	for range len(backends) {
		rr.current.Add(1)
		idx := rr.current.Load() % int32(len(backends))
		if backends[idx].IsAlive() {
			return backends[idx]
		}
	}

	return nil
}
