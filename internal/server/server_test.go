package server

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServer_StartAndShutdown(t *testing.T) {
	// Pick a random free port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	s := &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           http.NewServeMux(),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
	}()

	// Wait for the server to start
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
}

func TestServer_Shutdown_NoStart(t *testing.T) {
	s := &Server{
		httpServer: &http.Server{
			Addr:    "127.0.0.1:0",
			Handler: http.NewServeMux(),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Shutdown without Start should succeed
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
}
