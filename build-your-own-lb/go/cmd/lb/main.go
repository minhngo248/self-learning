package main

import (
	"fmt"
	log "log/slog"
	"net/http"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/algo"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/config"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/proxy"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		utils.Fatal("Failed to load config: %v", err)
	}

	// Backends
	backends := []string{
		"http://localhost:18000",
		"http://localhost:18001",
		"http://localhost:18002",
	}

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
		// init ALB
		lb = proxy.NewALB(lbAlgo)
		log.Info("Load balancer starting", "port", cfg.Port, "algorithm", cfg.LBAlgo)

		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), lb.(http.Handler)); err != nil {
			utils.Fatal("Server failed: %v", err)
		}
	default:
		utils.Fatal("Unsupported load balancer type: %s", cfg.LBType)
	}
}
