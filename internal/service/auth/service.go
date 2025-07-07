package auth

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ne4chelovek/auth-service/internal/cache"
	k "github.com/ne4chelovek/auth-service/internal/kafkaProducer"
	"github.com/ne4chelovek/auth-service/internal/repository"
	"github.com/ne4chelovek/auth-service/internal/utils"
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
	dbPool         *pgxpool.Pool
	blackList      cache.BlackList
	kafka          *k.Producer
	tokenUtils     utils.TokenUtils
}

func NewAuthService(userRepository repository.UsersRepository, dbPool *pgxpool.Pool, blackList cache.BlackList, kafka *k.Producer, tokenUtils utils.TokenUtils) *serv {
	return &serv{userRepository: userRepository, dbPool: dbPool, blackList: blackList, kafka: kafka, tokenUtils: tokenUtils}
}
