package proxy_test

import (
	"io"
	"net"
	"sync"
	"testing"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/algo"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/backend"
	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/proxy"
)

func TestNLB_handleConnectionWithLeastConn(t *testing.T) {
	b1 := backend.NewBackend("localhost:18000")
	b2 := backend.NewBackend("localhost:18001")
	b3 := backend.NewBackend("localhost:18002")

	backends := []*backend.Backend{
		b1, b2, b3,
	}

	lc := algo.NewLeastConn(backends)
	nlb := proxy.NewNLB(lc, backends)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	t.Cleanup(func() {
		listener.Close()
	})

	t.Logf("test listener: %s", listener.Addr().String())

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go nlb.HandleConnection(serverConn)

	if b1.NbConnection() != 0 {
		t.Fatalf("Connection is not closed properly")
	}
}

func TestNLB_NoKeepAliveClients(t *testing.T) {
	b1 := backend.NewBackend("localhost:18000")
	b2 := backend.NewBackend("localhost:18001")
	b3 := backend.NewBackend("localhost:18002")

	backends := []*backend.Backend{
		b1, b2, b3,
	}

	lc := algo.NewLeastConn(backends)
	nlb := proxy.NewNLB(lc, backends)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	t.Cleanup(func() {
		listener.Close()
	})

	go nlb.ListenAndServe(t.Context(), listener)

	var wg sync.WaitGroup
	for range 3 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			conn, err := net.Dial("tcp", listener.Addr().String())
			if err != nil {
				t.Errorf("failed to connect to NLB: %v", err)
				return
			}
			defer conn.Close()

			request := "GET / HTTP/1.1\r\n" +
				"Host: localhost\r\n" +
				"Connection: close\r\n" +
				"\r\n"

			if _, err := conn.Write([]byte(request)); err != nil {
				t.Errorf("failed to send request: %v", err)
				return
			}

			response, err := io.ReadAll(conn)
			if err != nil {
				t.Errorf("failed to read response: %v", err)
				return
			}

			if len(response) == 0 {
				t.Error("received empty response")
			}
		}()
	}

	wg.Wait()

}
