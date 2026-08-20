package algo_test

import (
	"fmt"
	"testing"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/algo"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
)

func TestRoundRobinNextAddr_AllHealthy(t *testing.T) {
	b1 := backend.NewBackend("http://localhost:18000")
	b2 := backend.NewBackend("http://localhost:18001")
	b3 := backend.NewBackend("http://localhost:18002")

	rr := algo.NewRoundRobin([]*backend.Backend{b1, b2, b3})

	if got := rr.NextAddr(); got != b1.Addr() {
		t.Fatalf("first backend = %q, want %q", got, b1.Addr())
	}
	if got := rr.NextAddr(); got != b2.Addr() {
		t.Fatalf("second backend = %q, want %q", got, b2.Addr())
	}
	if got := rr.NextAddr(); got != b3.Addr() {
		t.Fatalf("third backend = %q, want %q", got, b3.Addr())
	}
}

func TestRoundRobinNextAddr_AllUnhealthy(t *testing.T) {
	b1 := backend.NewBackend("http://localhost:18000")
	b1.SetNbRetry(4) // unhealthy
	b2 := backend.NewBackend("http://localhost:18001")
	b2.SetNbRetry(4) // unhealthy

	rr := algo.NewRoundRobin([]*backend.Backend{b1, b2})

	if got := rr.NextAddr(); got != "" {
		t.Fatalf("expected empty string for all unhealthy backends, got %q", got)
	}
}

func ExampleRoundRobin_NextAddr() {
	rr := algo.NewRoundRobin([]*backend.Backend{
		backend.NewBackend("http://localhost:18000"),
		backend.NewBackend("http://localhost:18001"),
		backend.NewBackend("http://localhost:18002"),
	})

	fmt.Println(rr.NextAddr())
	fmt.Println(rr.NextAddr())
	fmt.Println(rr.NextAddr())
	// Output:
	// http://localhost:18000
	// http://localhost:18001
	// http://localhost:18002
}
