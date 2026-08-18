package health

import (
	"context"
	log "log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
)

type ApplicationHealthChecker struct {
	backends      []*backend.Backend
	periodSeconds int
}

func NewApplicationHealthChecker(backends []*backend.Backend, periodSeconds int) *ApplicationHealthChecker {
	return &ApplicationHealthChecker{
		backends:      backends,
		periodSeconds: periodSeconds,
	}
}

func (ahc *ApplicationHealthChecker) CheckHealth(ctx context.Context, shutdownWg *sync.WaitGroup) {
	// Implement the health check logic for application load balancer
	// For example, you can send an HTTP request to the backend and check the response status code
	// Return true if the backend is healthy, false otherwise
	// This is a placeholder implementation, you should replace it with your actual health check logic
	ticker := time.NewTicker(time.Duration(ahc.periodSeconds) * time.Second)
	defer ticker.Stop()

	for t := range ticker.C {
		log.Info("Checking health of backends", "at", t)
		select {
		case <-ctx.Done():
			shutdownWg.Done()
			return
		default:
			addrToDelete := make([]string, 0)

			// Wait for all goroutines of THIS tick before reading NbRetry / filtering
			var tickWg sync.WaitGroup
			for _, be := range ahc.backends {
				tickWg.Add(1)
				go func(b *backend.Backend) {
					defer tickWg.Done()
					checkHealth(ctx, b)
				}(be)
			}
			tickWg.Wait() // guarantees no concurrent IncNbRetry when we read below

			for _, be := range ahc.backends {
				if be.IsUnhealthy() {
					addrToDelete = append(addrToDelete, be.Addr())
				}
			}

			toDelete := make(map[string]struct{}, len(addrToDelete))
			for _, addr := range addrToDelete {
				toDelete[addr] = struct{}{}
			}

			filteredBackends := ahc.backends[:0]
			for i := range ahc.backends {
				be := ahc.backends[i]
				if _, found := toDelete[be.Addr()]; !found {
					filteredBackends = append(filteredBackends, be)
				} else {
					log.Info("Removing unhealthy backend", "addr", be.Addr())
				}
			}
			ahc.backends = filteredBackends
		}
	}
}

func checkHealth(ctx context.Context, be *backend.Backend) {
	req, err := http.NewRequestWithContext(ctx, "GET", be.Addr()+"/healthz", nil)
	if err != nil {
		be.IncNbRetry()
		log.Warn("Failed to create health check request", "addr", be.Addr(), "error", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		be.IncNbRetry()
		log.Warn("Failed to check health", "addr", be.Addr(), "error", err, "time", be.NbRetry())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		be.IncNbRetry()
		log.Warn("Unhealthy backend", "addr", be.Addr(), "status", resp.StatusCode, "time", be.NbRetry())
		return
	}

	if be.IsRetrying() {
		log.Info("Backend recovered", "addr", be.Addr())
		be.ResetNbRetry()
	}
}
