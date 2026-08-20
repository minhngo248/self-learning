package algo

import (
	"math"
	"sync"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
)

type LeastConn struct {
	mu       sync.RWMutex
	backends []*backend.Backend
}

func NewLeastConn(backends []*backend.Backend) *LeastConn {
	return &LeastConn{
		mu:       sync.RWMutex{},
		backends: backends,
	}
}

func (lc *LeastConn) NextAddr() string {
	var minNbConn uint32
	minNbConn = math.MaxUint32
	var nextAddr string

	lc.mu.RLock()
	defer lc.mu.RUnlock()
	for _, backend := range lc.backends {
		if (backend.IsHealthy()) && (backend.NbConnection() < minNbConn) {
			minNbConn = backend.NbConnection()
			nextAddr = backend.Addr()
		}
	}

	return nextAddr
}
