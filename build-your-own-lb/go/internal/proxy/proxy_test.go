package proxy_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/algo"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/proxy"
)

func startHTTPBackends(t testing.TB, n int) ([]string, func()) {
	t.Helper()
	servers := make([]*httptest.Server, n)
	addrs := make([]string, n)

	for i := range n {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		servers[i] = srv
		addrs[i] = srv.URL
	}

	return addrs, func() {
		for _, srv := range servers {
			srv.Close()
		}
	}
}

func startTCPBackends(t testing.TB, n int) ([]string, func()) {
	t.Helper()
	listeners := make([]net.Listener, n)
	addrs := make([]string, n)

	for i := range n {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Failed to start TCP backend: %v", err)
		}
		listeners[i] = ln
		addrs[i] = ln.Addr().String()

		go func(l net.Listener) {
			for {
				conn, err := l.Accept()
				if err != nil {
					return
				}

				go func(c net.Conn) {
					defer c.Close()
					buf := make([]byte, 1024)
					c.Read(buf)
					c.Write([]byte("OK"))
				}(conn)
			}
		}(ln)
	}

	return addrs, func() {
		for _, ln := range listeners {
			ln.Close()
		}
	}
}

// ALB benchmark
func BenchmarkALB_ConnectionsPerSecond(b *testing.B) {
	addrs, cleanup := startHTTPBackends(b, 3)
	defer cleanup()

	backends := []*backend.Backend{
		backend.NewBackend(addrs[0]),
		backend.NewBackend(addrs[1]),
		backend.NewBackend(addrs[2]),
	}

	rr := algo.NewRoundRobin(backends)
	alb := proxy.NewALB(rr)

	srv := httptest.NewServer(alb)
	defer srv.Close()

	client := &http.Client{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(srv.URL + "/")
			if err != nil {
				b.Error(err)
				continue
			}
			resp.Body.Close()
		}
	})
}

// NLB benchmark
func BenchmarkNLB_ConnectionsPerSecond(b *testing.B) {
	addrs, cleanup := startTCPBackends(b, 3)
	defer cleanup()

	backends := []*backend.Backend{
		backend.NewBackend(addrs[0]),
		backend.NewBackend(addrs[1]),
		backend.NewBackend(addrs[2]),
	}

	rr := algo.NewRoundRobin(backends)
	nlb := proxy.NewNLB(rr)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	go nlb.ListenAndServe(ln)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, err := net.Dial("tcp", ln.Addr().String())
			if err != nil {
				b.Error(err)
				continue
			}
			conn.Write([]byte("ping"))
			buf := make([]byte, 2)
			conn.Read(buf)
			conn.Close()
		}
	})
}
