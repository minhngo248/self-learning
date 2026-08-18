package algo

import (
	log "log/slog"
	"sync/atomic"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
)

type RoundRobin struct {
	backends []*backend.Backend
	current  atomic.Uint64
}

// NewRoundRobin initializes the balancer with target backend addresses
func NewRoundRobin(backends []*backend.Backend) *RoundRobin {
	return &RoundRobin{
		backends: backends,
	}
}

// NextAddr retrieves the next backend in a thread-safe manner
func (rr *RoundRobin) NextAddr() string {
	for range rr.backends {
		idx := int(rr.current.Add(1)-1) % len(rr.backends)
		be := rr.backends[idx]
		if be.IsHealthy() {
			return be.Addr()
		}
	}

	log.Warn("All backends degraded")
	return "" // all backends degraded
}
