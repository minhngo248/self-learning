package proxy

import (
	log "log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/algo"
)

type ALB struct {
	lbAlgo algo.LBAlgo
}

func NewALB(lbAlgo algo.LBAlgo) *ALB {
	return &ALB{
		lbAlgo: lbAlgo,
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
	log.Info("Request routed to", "host", targetURL.Host)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(w, r)
}
