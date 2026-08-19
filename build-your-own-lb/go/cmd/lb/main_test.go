package main_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/algo"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/health"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/proxy"
)

func TestMain_CallTheUnhealthyBackendInTheFirstPeriodSecond(t *testing.T) {
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("healthy"))
	}))
	defer healthyServer.Close()

	unhealthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("backend is unhealthy"))
	}))
	defer unhealthyServer.Close()

	// Place the backend that will be chosen first on the first request
	// before the health checker has had time to mark it unhealthy.
	unhealthy := backend.NewBackend(unhealthyServer.URL)
	healthy := backend.NewBackend(healthyServer.URL)
	backends := []*backend.Backend{unhealthy, healthy}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var shutdownWg sync.WaitGroup
	shutdownWg.Add(1)

	periodSeconds := 10
	healthChecker := health.NewApplicationHealthChecker(backends, periodSeconds)
	go healthChecker.CheckHealth(ctx, &shutdownWg, healthChecker.CustomCheckHealth)

	<-healthChecker.Ready()

	lbAlgo := algo.NewRoundRobin(backends)
	alb := proxy.NewALB(lbAlgo)
	server := httptest.NewServer(alb)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK before the first health-check window completes, got %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}
