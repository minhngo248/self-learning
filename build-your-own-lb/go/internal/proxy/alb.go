package proxy

import (
	"context"
	"errors"
	log "log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/algo"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
)

type ALB struct {
	lbAlgo   algo.LBAlgo
	backends []*backend.Backend
}

func NewALB(lbAlgo algo.LBAlgo, backends []*backend.Backend) *ALB {
	return &ALB{
		lbAlgo:   lbAlgo,
		backends: backends,
	}
}

func (alb *ALB) ListenAndServe(ctx context.Context, listener net.Listener) error {
	server := &http.Server{
		Handler: alb,
	}

	go func() {
		<-ctx.Done()
		//log.Info("ALB shutting down")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("Failed to shut down ALB", "error", err)
		}
	}()

	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// ServeHTTP satisfies the http.Handler interface for http.ListenAndServe
func (alb *ALB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	targetAddr := alb.lbAlgo.NextAddr()
	if targetAddr == "" {
		http.Error(w, "No backends available", http.StatusServiceUnavailable)
		return
	}

	targetURL, err := url.Parse(targetAddr)
	if err != nil {
		http.Error(w, "Invalid backend URL", http.StatusInternalServerError)
		return
	}

	// ReverseProxy forwards method, headers, query params, body, and context
	backend := backend.BackendByAddr(alb.backends, targetAddr)
	backend.IncNbConnection()
	defer backend.DecNbConnection()

	log.Debug("Request routed to", "host", targetURL.Host)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(w, r)
}
