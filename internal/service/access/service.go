package access

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ne4chelovek/auth-service/internal/cache"
	"github.com/ne4chelovek/auth-service/internal/repository"
	"github.com/ne4chelovek/auth-service/internal/utils"
	"time"
)

const (
	accessTokenExpiration    = 30 * time.Minute
	accessTokenSecretKeyName = "access"
)

type serv struct {
	accessRepository repository.AccessRepository
	blackList        cache.BlackList
	dbPool           *pgxpool.Pool
	tokenUtils       utils.TokenUtils
}

func NewAccessService(accessRepository repository.AccessRepository, blackList cache.BlackList, dbPool *pgxpool.Pool, tokenUtils utils.TokenUtils) *serv {
	return &serv{accessRepository: accessRepository, blackList: blackList, dbPool: dbPool, tokenUtils: tokenUtils}
}
