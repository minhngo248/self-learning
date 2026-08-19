package proxy

import (
	"io"
	log "log/slog"
	"net"
	"sync"

	"github.com/minhngo248/self-learning/build-your-own-lb/go/internal/algo"
)

type NLB struct {
	lbAlgo algo.LBAlgo
}

func NewNLB(lbAlgo algo.LBAlgo) *NLB {
	return &NLB{
		lbAlgo: lbAlgo,
	}
}

func (nlb *NLB) ListenAndServe(listener net.Listener) {
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Failed to accept connection", "error", err)
			continue
		}

		go nlb.handleConnection(conn)
	}
}

func (nlb *NLB) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Pick a backend
	backendAddr := nlb.lbAlgo.NextAddr()
	if backendAddr == "" {
		log.Error("No healthy backend available")
		return
	}

	// Connect to the backend
	backendConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		log.Error("Failed to connect to backend", "backend", backendAddr, "error", err)
		return
	}
	defer backendConn.Close()

	// Pipe bytes in both directions concurrently
	var wg sync.WaitGroup
	wg.Add(2)

	// client to backend
	go func() {
		defer wg.Done()
		_, err := io.Copy(backendConn, conn)
		if err != nil {
			log.Error("Error while copying from client to backend", "error", err)
		}
		//log.Info("Finished copying from client to backend", "backend", backendAddr)
		// signal backend we're done sending
		if tcpConn, ok := backendConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	// backend to client
	go func() {
		defer wg.Done()
		_, err := io.Copy(conn, backendConn)
		if err != nil {
			log.Error("Error while copying from backend to client", "error", err)
		}
		//log.Info("Finished copying from backend to client", "backend", backendAddr)
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	wg.Wait()
}
