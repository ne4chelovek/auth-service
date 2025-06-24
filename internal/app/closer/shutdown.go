package closer

import (
	"context"
	server "github.com/ne4chelovek/auth-service/internal/app"
	"github.com/ne4chelovek/auth-service/internal/logger"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func WaitForShutdown(ctx context.Context, errChan <-chan error, s *server.Servers) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		logger.Info("Received shutdown signal")
	case err := <-errChan:
		logger.Error("Critical error: ", zap.Error(err))
	case <-ctx.Done():
		logger.Info("Context cancelled")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	logger.Info("Stopping gRPC server...")
	s.GRPC.GracefulStop()

	logger.Info("Stopping HTTP server...")
	if err := s.HTTP.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error:", zap.Error(err))
	}

	logger.Info("Stopping Swagger...")
	if err := s.Swagger.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error:", zap.Error(err))
	}

	logger.Info("Stopping Prometheus...")
	if err := s.Prometheus.Shutdown(shutdownCtx); err != nil {
		logger.Error("Prometheus shutdown error:", zap.Error(err))
	}

	logger.Info("Closing database connections...")
	s.DB.Close()
}
