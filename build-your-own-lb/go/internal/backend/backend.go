package backend

import (
	"sync/atomic"
)

type Backend struct {
	addr         string // for roundrobin
	nbRetry      atomic.Uint32
	nbConnection atomic.Uint32 // for leastconn
}

func NewBackend(addr string) *Backend {
	return &Backend{
		addr:         addr,
		nbRetry:      atomic.Uint32{},
		nbConnection: atomic.Uint32{},
	}
}

func BackendByAddr(backends []*Backend, addr string) *Backend {
	for _, backend := range backends {
		if backend.Addr() == addr {
			return backend
		}
	}
	return nil
}

func (be *Backend) Addr() string {
	return be.addr
}

func (be *Backend) NbRetry() uint32 {
	return be.nbRetry.Load()
}

func (be *Backend) SetNbRetry(nbRetry uint32) {
	be.nbRetry.Store(nbRetry)
}

func (be *Backend) IncNbRetry() {
	be.nbRetry.Add(1)
}

func (be *Backend) ResetNbRetry() {
	be.nbRetry.Store(0)
}

func (be *Backend) IsHealthy() bool {
	return be.NbRetry() == 0
}

func (be *Backend) IsRetrying() bool {
	return be.NbRetry() > 0 && be.NbRetry() <= 3
}

func (be *Backend) IsUnhealthy() bool {
	return be.NbRetry() > 3
}

func (be *Backend) NbConnection() uint32 {
	return be.nbConnection.Load()
}

func (be *Backend) IncNbConnection() {
	be.nbConnection.Add(1)
}

func (be *Backend) DecNbConnection() {
	be.nbConnection.Add(^uint32(0)) // Decrement by 1
}

func (be *Backend) ResetNbConnection() {
	be.nbConnection.Store(0)
}
