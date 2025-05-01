package utils

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/ne4chelovek/auth-service/internal/model"
	"time"
)

func GenerateToken(info model.UserInfo, secretKey []byte, duration time.Duration) (string, error) {
	claims := model.UserClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(duration).Unix(),
		},
		info.Username,
		info.Password,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}

func VerifyToken(tokenStr string, secretKey []byte) (*model.UserClaims, error) {
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
		return nil, fmt.Errorf("invalid token: %s", err.Error())
	}
	claims, ok := token.Claims.(*model.UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
