package auth

import (
	"context"
	"github.com/ne4chelovek/auth-service/internal/model"
	"github.com/ne4chelovek/auth-service/internal/utils"
)

func (s *serv) GetAccessToken(ctx context.Context, token string) (string, error) {
	claims, err := utils.VerifyToken(token, []byte(refreshTokenSecretKeyName))
	if err != nil {
		return "", err
	}

	accessToken, err := utils.GenerateToken(model.UserInfo{
		Username: claims.Username,
		Role:     claims.Role,
	},
		[]byte(accessTokenSecretKeyName),
		accessTokenExpiration,
	)
	if err != nil {
		return "", err
	}

	return accessToken, nil

}
