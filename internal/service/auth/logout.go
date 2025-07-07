package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ne4chelovek/auth-service/internal/logger"
	"go.uber.org/zap"
	"time"
)

func (s *serv) Logout(ctx context.Context, accessToken string) error {
	claims, err := s.tokenUtils.VerifyToken(ctx, accessToken, []byte(accessTokenSecretKeyName))
	if err != nil {
		return fmt.Errorf("logout: invalid access token %v", err)
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
