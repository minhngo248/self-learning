package health

import (
	"context"
	log "log/slog"
	"net"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
)

type NetworkHealthChecker struct {
	HealthCheckerImpl
}

func NewNetworkHealthChecker(backends []*backend.Backend, periodSeconds int) *NetworkHealthChecker {
	return &NetworkHealthChecker{
		HealthCheckerImpl: *NewHealthChecker(backends, periodSeconds),
	}
}

func (nhc *NetworkHealthChecker) CustomCheckHealth(ctx context.Context, be *backend.Backend) {
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", be.Addr())
	if err != nil {
		be.IncNbRetry()
		log.Error("Failed to connect to backend", "addr", be.Addr(), "error", err, "time", be.NbRetry())
		return
	}
	defer conn.Close()

	if be.IsRetrying() {
		log.Info("Backend recovered", "addr", be.Addr())
		be.ResetNbRetry()
	}
}
