package main

import (
	"context"
	"fmt"
	server "github.com/ne4chelovek/auth-service/internal/app"
	"github.com/ne4chelovek/auth-service/internal/app/closer"
	"log"
	"net/http"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	servers, err := server.SetupServers(ctx)
	if err != nil {
		log.Fatalf("Failed to setup servers: %v", err)
	}

	errChan := make(chan error, 3)

	go runGRPCServer(servers, errChan)
	go runHTTPServer(servers.HTTP, errChan)
	go runSwaggerServer(servers.Swagger, errChan)

	closer.WaitForShutdown(ctx, cancel, errChan, servers)
}

func runGRPCServer(s *server.Servers, errChan chan<- error) {
	log.Printf("Starting gRPC server on %s", s.Listener.Addr())
	if err := s.GRPC.Serve(s.Listener); err != nil {
		errChan <- fmt.Errorf("gRPC server error: %w", err)
	}
}

func runHTTPServer(s *http.Server, errChan chan<- error) {
	log.Printf("Starting HTTP server on %s", s.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		errChan <- fmt.Errorf("HTTP server error: %w", err)
	}
}

func runSwaggerServer(s *http.Server, errChan chan<- error) {
	log.Printf("Starting Swagger server on %s", s.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		errChan <- fmt.Errorf("Swagger server error: %w", err)
	}
}
