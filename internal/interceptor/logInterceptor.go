package interceptor

import (
	"context"
	"github.com/ne4chelovek/auth-service/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

func LogInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	now := time.Now()

	res, err := handler(ctx, req)
	if err != nil {
		logger.Error(err.Error(), zap.String("method", info.FullMethod), zap.Any("request", req))
	}
	logger.Info("request", zap.String("method", info.FullMethod), zap.Any("request", req), zap.Any("response", res), zap.Duration("duration", time.Since(now)))

	return res, err
}
