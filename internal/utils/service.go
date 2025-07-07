package utils

import (
	"context"
	"github.com/ne4chelovek/auth-service/internal/model"
	"time"
)

type TokenUtils interface {
	GenerateToken(info model.UserInfo, secretKey []byte, duration time.Duration) (string, error)
	VerifyToken(ctx context.Context, tokenStr string, secretKey []byte) (*model.UserClaims, error)
}
