package utils

import (
	"context"
	"errors"
	"fmt"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/ne4chelovek/auth-service/internal/cache"
	"github.com/ne4chelovek/auth-service/internal/model"
	"github.com/redis/go-redis/v9"
	"time"
)

type blackListToken struct {
	blackList cache.BlackList
}

func NewTokenService(blackList cache.BlackList) *blackListToken {
	return &blackListToken{blackList: blackList}
}

func (b *blackListToken) GenerateToken(info model.UserInfo, secretKey []byte, duration time.Duration) (string, error) {
	claims := model.UserClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
		info.Username,
		info.Role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}

func (b *blackListToken) VerifyToken(ctx context.Context, tokenStr string, secretKey []byte) (*model.UserClaims, error) {
	_, err := b.blackList.Get(ctx, tokenStr)
	switch {
	case err == nil:
		return nil, fmt.Errorf("token is invalidated")
	case !errors.Is(err, redis.Nil):
		return nil, fmt.Errorf("failed to check token validity: %w", err)
	}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&model.UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return secretKey, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(*model.UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
