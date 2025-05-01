package auth

import (
	"context"
	"fmt"
	"github.com/ne4chelovek/auth-service/internal/model"
	"github.com/ne4chelovek/auth-service/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func (s *serv) Login(ctx context.Context, login *model.UserCreds) (string, error) {
	authInfo, err := s.userRepository.GetAuthInfo(ctx, login.UserNames)
	if err != nil {
		return "", fmt.Errorf("User not found: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(authInfo.Password), []byte(login.Password))
	if err != nil {
		return "", fmt.Errorf("Password incorrect: %v", err)
	}

	refreshToken, err := utils.GenerateToken(model.UserInfo{
		Username: authInfo.Username,
		Role:     authInfo.Role,
	},
		[]byte(refreshTokenSecretKeyName),
		refreshTokenExpiration,
	)
	if err != nil {
		return "", err
	}
	return refreshToken, nil
}
