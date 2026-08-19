package health

import (
	"context"
	log "log/slog"
	"net/http"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
)

type ApplicationHealthChecker struct {
	HealthCheckerImpl
}

func NewApplicationHealthChecker(backends []*backend.Backend, periodSeconds int) *ApplicationHealthChecker {
	return &ApplicationHealthChecker{
		HealthCheckerImpl: *NewHealthChecker(backends, periodSeconds),
	}
}

func (ahc *ApplicationHealthChecker) CustomCheckHealth(ctx context.Context, be *backend.Backend) {
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
