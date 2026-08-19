package main

import (
	"context"
	"fmt"
	log "log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/algo"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/config"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/health"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/proxy"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		utils.Fatal("Failed to load config: %v", err)
	}

	// Backends
	backendAddrs := []string{
		"localhost:18000",
		"localhost:18001",
		"localhost:18002",
	}

	// Set default logger level
	config.InitLogger(cfg.Profile)

	// Init backends
	backends := make([]*backend.Backend, len(backendAddrs))
	for i, backendAddr := range backendAddrs {
		backends[i] = backend.NewBackend(backendAddr)
	}

	var shutdownWg sync.WaitGroup
	shutdownWg.Add(1)

	// Start background context for health checking
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// health checker properties
	periodSeconds := 10

	var lbAlgo algo.LBAlgo
	switch cfg.LBAlgo {
	case "roundrobin":
		// init round robin algorithm
		lbAlgo = algo.NewRoundRobin(backends)
	default:
		utils.Fatal("Unsupported load balancing algorithm: %s", cfg.LBAlgo)
	}

	var lb any
	switch cfg.LBType {
	case "application":
		// Init application health checker
		healthChecker := health.NewApplicationHealthChecker(backends, periodSeconds)
		go healthChecker.CheckHealth(ctx, &shutdownWg, healthChecker.CustomCheckHealth)

		<-healthChecker.Ready()

		// init ALB
		lb = proxy.NewALB(lbAlgo)
		log.Info("Load balancer starting", "port", cfg.Port, "algorithm", cfg.LBAlgo)

		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), lb.(http.Handler)); err != nil {
			utils.Fatal("Server failed: %v", err)
		}
	case "network":
		// Init network health checker
		healthChecker := health.NewNetworkHealthChecker(backends, periodSeconds)
		go healthChecker.CheckHealth(ctx, &shutdownWg, healthChecker.CustomCheckHealth)

		<-healthChecker.Ready()

		// init NLB
		lb = proxy.NewNLB(lbAlgo)
		log.Info("Load balancer starting", "port", cfg.Port, "algorithm", cfg.LBAlgo)

		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
		if err != nil {
			utils.Fatal("Failed to start listener: %v", err)
		}

		lb.(*proxy.NLB).ListenAndServe(listener)
	default:
		utils.Fatal("Unsupported load balancer type: %s", cfg.LBType)
	}
}
