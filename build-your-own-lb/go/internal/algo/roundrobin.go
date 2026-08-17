package algo

import (
	"sync/atomic"
)

type RoundRobin struct {
	backends []string
	incr     atomic.Uint64
}

// NewRoundRobin initializes the balancer with target backend addresses
func NewRoundRobin(backends []string) *RoundRobin {
	return &RoundRobin{
		backends: backends,
	}
}

// NextAddr retrieves the next backend in a thread-safe manner
func (rr *RoundRobin) NextAddr() string {
	if len(rr.backends) == 0 {
		return ""
	}
	n := rr.incr.Add(1) - 1
	return rr.backends[n%uint64(len(rr.backends))]
}
