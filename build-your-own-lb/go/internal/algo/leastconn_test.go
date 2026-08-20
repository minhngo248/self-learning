package algo_test

import (
	"testing"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/algo"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
)

func TestLeastConn_NextAddrAllHealthy(t *testing.T) {
	b1 := backend.NewBackend("localhost:18000")
	b2 := backend.NewBackend("localhost:18001")
	b3 := backend.NewBackend("localhost:18002")

	backends := []*backend.Backend{
		b1, b2, b3,
	}

	lc := algo.NewLeastConn(backends)

	if got := lc.NextAddr(); got != "localhost:18000" {
		t.Fatalf("first backend = %q, want %q", got, "localhost:18000")
	}

	b1.IncNbConnection()

	if got := lc.NextAddr(); got != "localhost:18001" {
		t.Fatalf("second backend = %q, want %q", got, "localhost:18001")
	}
}
