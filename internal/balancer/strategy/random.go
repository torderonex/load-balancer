// internal/balancer/strategy/random.go
package strategy

import (
	"log/slog"
	"math/rand"
	"time"

	"github.com/torderonex/load-balancer/internal/model"
)

type randomStrategy struct {
	rand *rand.Rand
}

func NewRandomStrategy() Strategy {
	source := rand.NewSource(time.Now().UnixNano())
	return &randomStrategy{
		rand: rand.New(source),
	}
}

func (rs *randomStrategy) NextBackend(backends []*model.Backend) *model.Backend {
	if len(backends) == 0 {
		return nil
	}

	aliveBackends := make([]*model.Backend, 0, len(backends))
	for _, backend := range backends {
		if backend.IsAlive() {
			aliveBackends = append(aliveBackends, backend)
		}
	}

	if len(aliveBackends) == 0 {
		return nil
	}

	idx := rs.rand.Intn(len(aliveBackends))
	slog.Debug("Random strategy selected backend", "index", idx, "total", len(aliveBackends))
	return aliveBackends[idx]
}
