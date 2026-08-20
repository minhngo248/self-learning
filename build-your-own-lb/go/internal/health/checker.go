package health

import (
	"context"
	log "log/slog"
	"sync"
	"time"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
)

type HealthChecker interface {
	Ready() <-chan struct{}
	CheckHealth(ctx context.Context, shutdownWg *sync.WaitGroup, customCheckHealth func(ctx context.Context, be *backend.Backend))
}

type HealthCheckerImpl struct {
	mu            sync.RWMutex
	backends      []*backend.Backend
	periodSeconds int
	ready         chan struct{}
	readyOnce     sync.Once
}

func NewHealthChecker(backends []*backend.Backend, periodSeconds int) *HealthCheckerImpl {
	return &HealthCheckerImpl{
		mu:            sync.RWMutex{},
		backends:      backends,
		periodSeconds: periodSeconds,
		ready:         make(chan struct{}),
	}
}

func (hc *HealthCheckerImpl) Ready() <-chan struct{} {
	return hc.ready
}

func (hc *HealthCheckerImpl) CheckHealth(ctx context.Context, shutdownWg *sync.WaitGroup, customCheckHealth func(ctx context.Context, be *backend.Backend)) {
	ticker := time.NewTicker(time.Duration(hc.periodSeconds) * time.Second)
	defer ticker.Stop()

	for t := range ticker.C {
		log.Debug("Checking health of backends", "at", t)
		select {
		case <-ctx.Done():
			shutdownWg.Done()
			return
		default:
			addrToDelete := make([]string, 0)

			// Wait for all goroutines of THIS tick before reading NbRetry / filtering
			var tickWg sync.WaitGroup
			for _, be := range hc.backends {
				tickWg.Add(1)
				go func(b *backend.Backend) {
					defer tickWg.Done()
					customCheckHealth(ctx, b)
				}(be)
			}
			tickWg.Wait() // guarantees no concurrent IncNbRetry when we read below
			hc.readyOnce.Do(func() {
				close(hc.ready)
			})

			for _, be := range hc.backends {
				if be.IsUnhealthy() {
					addrToDelete = append(addrToDelete, be.Addr())
				}
			}

			toDelete := make(map[string]struct{}, len(addrToDelete))
			for _, addr := range addrToDelete {
				toDelete[addr] = struct{}{}
			}

			filteredBackends := make([]*backend.Backend, 0, len(hc.backends)-len(toDelete))
			for i := range hc.backends {
				be := hc.backends[i]
				if _, found := toDelete[be.Addr()]; !found {
					filteredBackends = append(filteredBackends, be)
				} else {
					log.Info("Removing unhealthy backend", "addr", be.Addr())
				}
			}
			hc.mu.Lock()
			hc.backends = filteredBackends
			hc.mu.Unlock()
		}
	}
}

func (hc *HealthCheckerImpl) Backends() []*backend.Backend {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.backends
}
