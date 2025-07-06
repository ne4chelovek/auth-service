package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ne4chelovek/auth-service/internal/logger"
	"github.com/ne4chelovek/auth-service/internal/utils"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

func (s *serv) Logout(ctx context.Context, accessToken string) error {
	_, err := s.blackList.Get(ctx, accessToken)
	switch {
	case err == nil:
		return fmt.Errorf("logout: token already invalidated")
	case !errors.Is(err, redis.Nil):
		return fmt.Errorf("logout: failed to check token: %w", err)
	}

	claims, err := utils.VerifyToken(accessToken, []byte(accessTokenSecretKeyName))
	if err != nil {
		return fmt.Errorf("logout: invalid refresh token")
	}

	remainingTTL := time.Until(claims.ExpiresAt.Time)
	if err := s.blackList.BlackListToken(ctx, accessToken, remainingTTL); err != nil {
		return fmt.Errorf("logout: failed to black list")
	}

	event := map[string]interface{}{
		"event_type": "user_logout",
		"user_id":    claims.Username,
		"session_id": accessToken,
		"timestamp":  time.Now().Unix(),
	}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal logout event",
			zap.String("user_id", claims.Username),
			zap.Error(err))
		return fmt.Errorf("logout: failed to prepare event: %w", err)
	}

	if err := s.kafka.Produce(eventBytes, "user_session_events"); err != nil {
		logger.Error("Failed to send logout event to Kafka",
			zap.String("user_id", claims.Username),
			zap.Error(err))
	}

	return nil
}
