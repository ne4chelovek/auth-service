package interceptor

import (
	"context"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CircuitBreaker struct {
	cd *gobreaker.CircuitBreaker
}

func NewCircuitBreakerInterceptor(cd *gobreaker.CircuitBreaker) *CircuitBreaker {
	return &CircuitBreaker{cd: cd}
}

func (c *CircuitBreaker) Unary(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	res, err := c.cd.Execute(func() (interface{}, error) {
		return handler(ctx, request)
	})
	if err != nil {
		if err == gobreaker.ErrOpenState {
			return nil, status.Error(codes.Unavailable, "service is busy")
		}
		return nil, err
	}
	return res, nil
}
