package strategy

import (
	"fmt"
	"log/slog"

	"github.com/torderonex/load-balancer/internal/model"
	"github.com/torderonex/load-balancer/pkg/sl"
)

type Strategy interface {
	NextBackend(backends []*model.Backend) *model.Backend
}

func NewStrategy(strategy string) Strategy {
	switch strategy {
	case "round_robin":
		return NewRoundRobinStrategy()
	case "random":
		return NewRandomStrategy()
	default:
		slog.Error("Unknown strategy, using round_robin", sl.Err(fmt.Errorf("unknown strategy: %s", strategy)))
		return NewRoundRobinStrategy()
	}
}
