package health

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
)

type blockingTransport struct {
	started chan struct{}
	release chan struct{}
}

func (bt *blockingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	bt.started <- struct{}{}
	<-bt.release

	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Status:     "500 Internal Server Error",
		Body:       io.NopCloser(strings.NewReader("error")),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestCheckHealth_WaitsForAllHealthChecksBeforeFiltering(t *testing.T) {
	b1 := backend.NewBackend("http://localhost:18000")
	b2 := backend.NewBackend("http://localhost:18001")
	b1.SetNbRetry(3)
	b2.SetNbRetry(3)

	oldTransport := http.DefaultClient.Transport
	tr := &blockingTransport{
		started: make(chan struct{}, 2),
		release: make(chan struct{}),
	}
	http.DefaultClient.Transport = tr
	defer func() {
		http.DefaultClient.Transport = oldTransport
	}()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var shutdownWg sync.WaitGroup
	shutdownWg.Add(1)

	checker := NewApplicationHealthChecker([]*backend.Backend{b1, b2}, 1)
	go checker.CheckHealth(ctx, &shutdownWg, checker.CustomCheckHealth)

	<-tr.started
	<-tr.started

	if b1.NbRetry() != 3 || b2.NbRetry() != 3 {
		t.Fatalf("health checks should still be waiting; retries are %d and %d", b1.NbRetry(), b2.NbRetry())
	}
	if len(checker.backends) != 2 {
		t.Fatalf("filtering happened before all health checks finished; len(backends) = %d", len(checker.backends))
	}

	close(tr.release)

	time.Sleep(1500 * time.Millisecond)
	if len(checker.backends) != 0 {
		t.Fatalf("expected both backends to be filtered after retries complete, got %d backends", len(checker.backends))
	}

	cancel()
	shutdownWg.Wait()
}
