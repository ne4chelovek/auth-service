package blackList

import (
	"context"
	"github.com/ne4chelovek/auth-service/internal/cache"
	"github.com/redis/go-redis/v9"
	"time"
)

type clientRedis struct {
	redisClient *redis.Client
}

func NewBlackList(redisClient *redis.Client) cache.BlackListRepository {
	return &clientRedis{redisClient: redisClient}
}

func (r *clientRedis) BlackListToken(ctx context.Context, token string, ttl time.Duration) error {
	return r.redisClient.Set(ctx, "blacklist:"+token, "1", ttl).Err()
}

func (r *clientRedis) Get(ctx context.Context, token string) (string, error) {
	return r.redisClient.Get(ctx, "blacklist:"+token).Result()
}
