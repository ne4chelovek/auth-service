package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ne4chelovek/auth-service/internal/logger"
	"github.com/ne4chelovek/auth-service/internal/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func (s *serv) Login(ctx context.Context, username, password string) (string, error) {
	authInfo, err := s.userRepository.GetAuthInfo(ctx, username)
	if err != nil {
		return "", fmt.Errorf("user not found: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(authInfo.Password), []byte(password))
	if err != nil {
		return "", fmt.Errorf("password incorrect: %v", err)
	}

	refreshToken, err := s.tokenUtils.GenerateToken(model.UserInfo{
		Username: authInfo.Username,
		Role:     authInfo.Role,
	},
		[]byte(refreshTokenSecretKeyName),
		refreshTokenExpiration,
	)
	if err != nil {
		return "", err
	}

	event := map[string]interface{}{
		"event_type": "user_login",
		"user_id":    authInfo.Username,
		"session_id": refreshToken,
		"role":       authInfo.Role,
		"timestamp":  time.Now().Unix(),
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return "", err
	}

	if err = s.kafka.Produce(eventBytes, "user_session_events"); err != nil {
		logger.Error("Failed to send login event to Kafka",
			zap.String("user_name", username),
			zap.Error(err))
	}

	return refreshToken, nil
}
