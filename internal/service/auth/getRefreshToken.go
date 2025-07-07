package auth

import (
	"context"
	"fmt"
	"github.com/ne4chelovek/auth-service/internal/model"
)

func (s *serv) GetRefreshToken(ctx context.Context, oldRefreshToken string) (string, error) {
	claims, err := s.tokenUtils.VerifyToken(ctx, oldRefreshToken, []byte(refreshTokenSecretKeyName))
	if err != nil {
		return "", fmt.Errorf("Invalid refresh token: %v", err)
	}

	refreshToken, err := s.tokenUtils.GenerateToken(model.UserInfo{
		Username: claims.Username,
		Role:     claims.Role,
	},
		[]byte(refreshTokenSecretKeyName),
		refreshTokenExpiration,
	)

	if err != nil {
		return "", fmt.Errorf("Failed to generate refresh token: %v", err)
	}

	return refreshToken, nil
}
