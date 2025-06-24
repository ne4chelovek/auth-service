package cache

import (
	"context"
	"time"
)

type BlackListRepository interface {
	BlackListToken(ctx context.Context, token string, ttl time.Duration) error
	Get(ctx context.Context, token string) (string, error)
}
