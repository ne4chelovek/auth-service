package auth

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ne4chelovek/auth-service/internal/repository"
	"time"
)

const (
	refreshTokenExpiration    = 360 * time.Minute
	accessTokenExpiration     = 30 * time.Minute
	refreshTokenSecretKeyName = "refresh"
	accessTokenSecretKeyName  = "access"
)

type serv struct {
	userRepository repository.UsersRepository

	dbPool *pgxpool.Pool
}

func NewAuthService(userRepository repository.UsersRepository, dbPool *pgxpool.Pool) *serv {
	return &serv{userRepository: userRepository, dbPool: dbPool}
}
