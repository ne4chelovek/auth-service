package main

import (
	"context"
	"errors"
	"fmt"
	server "github.com/ne4chelovek/auth-service/internal/app"
	"github.com/ne4chelovek/auth-service/internal/app/closer"
	"github.com/ne4chelovek/auth-service/internal/logger"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	servers, err := server.SetupServers(ctx)
	if err != nil {
		logger.Fatal("Failed to setup servers: %v", zap.Error(err))
	}

	errChan := make(chan error, 1)

	go runGRPCServer(servers, errChan)
	go runHTTPServer(servers.HTTP, errChan)
	go runSwaggerServer(servers.Swagger, errChan)
	go runPrometheusServer(servers.Prometheus, errChan)

	closer.WaitForShutdown(ctx, errChan, servers)
}

func runGRPCServer(s *server.Servers, errChan chan<- error) {
	logger.Info("Starting gRPC server on ", zap.String("address", s.Listener.Addr().String()))
	if err := s.GRPC.Serve(s.Listener); err != nil {
		errChan <- fmt.Errorf("gRPC server error: %w", err)
	}
}

func runHTTPServer(s *http.Server, errChan chan<- error) {
	log.Printf("Starting HTTP server on %s", s.Addr)
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		errChan <- fmt.Errorf("HTTP server error: %w", err)
	}
}

func runSwaggerServer(s *http.Server, errChan chan<- error) {
	log.Printf("Starting Swagger server on %s", s.Addr)
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		errChan <- fmt.Errorf("swagger server error: %w", err)
	}
}

func runPrometheusServer(s *http.Server, errChan chan<- error) {
	log.Printf("Starting Prometheus server on %s", s.Addr)
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		errChan <- fmt.Errorf("prometheus server error: %w", err)
	}
}
