package main

import (
	"context"
	"fmt"
	log "log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	healthCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// health checker properties
	periodSeconds := 10

	var lbAlgo algo.LBAlgo
	switch cfg.LBAlgo {
	case "roundrobin":
		// init round robin algorithm
		lbAlgo = algo.NewRoundRobin(backends)
	case "leastconn":
		// init least connection algorithm
		lbAlgo = algo.NewLeastConn(backends)
	default:
		utils.Fatal("Unsupported load balancing algorithm: %s", cfg.LBAlgo)
	}

	lbCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	var lb proxy.LB
	switch cfg.LBType {
	case "application":
		// Init application health checker
		healthChecker := health.NewApplicationHealthChecker(backends, periodSeconds)
		go healthChecker.CheckHealth(healthCtx, &shutdownWg, healthChecker.CustomCheckHealth)

		<-healthChecker.Ready()

		// init ALB
		lb = proxy.NewALB(lbAlgo, backends)
		log.Info("Load balancer starting", "port", cfg.Port, "algorithm", cfg.LBAlgo)

		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
		if err != nil {
			utils.Fatal("Failed to start listener: %v", err)
		}

		if err := lb.(*proxy.ALB).ListenAndServe(lbCtx, listener); err != nil {
			utils.Fatal("Server failed: %v", err)
		}
	case "network":
		// Init network health checker
		healthChecker := health.NewNetworkHealthChecker(backends, periodSeconds)
		go healthChecker.CheckHealth(healthCtx, &shutdownWg, healthChecker.CustomCheckHealth)

		<-healthChecker.Ready()

		// init NLB
		lb = proxy.NewNLB(lbAlgo, backends)
		log.Info("Load balancer starting", "port", cfg.Port, "algorithm", cfg.LBAlgo)

		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
		if err != nil {
			utils.Fatal("Failed to start listener: %v", err)
		}

		lb.(*proxy.NLB).ListenAndServe(lbCtx, listener)
	default:
		utils.Fatal("Unsupported load balancer type: %s", cfg.LBType)
	}
}
