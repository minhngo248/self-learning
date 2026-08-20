package proxy

import (
	log "log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

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
