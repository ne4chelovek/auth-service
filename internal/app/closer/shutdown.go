package closer

import (
	"context"
	server "github.com/ne4chelovek/auth-service/internal/app"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func WaitForShutdown(ctx context.Context, cancel context.CancelFunc, errChan <-chan error, s *server.Servers) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("Received shutdown signal")
	case err := <-errChan:
		log.Printf("Critical error: %v", err)
	case <-ctx.Done():
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// 1. Остановка gRPC сервера
	log.Println("Stopping gRPC server...")
	s.GRPC.GracefulStop()
	// 2. Остановка HTTP Gateway
	log.Println("Stopping HTTP server...")
	if err := s.HTTP.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	// 3. Остановка Swagger UI
	log.Println("Stopping Swagger UI...")
	if err := s.Swagger.Shutdown(shutdownCtx); err != nil {
		log.Printf("Swagger server shutdown error: %v", err)
	}
	// 4. Закрытие соединения с БД
	log.Println("Closing database connections...")
	s.DB.Close()
	cancel()
}
