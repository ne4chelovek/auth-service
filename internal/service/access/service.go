package access

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ne4chelovek/auth-service/internal/cache"
	"github.com/ne4chelovek/auth-service/internal/repository"
	"time"
)

const (
	accessTokenExpiration    = 30 * time.Minute
	accessTokenSecretKeyName = "access"
)

type serv struct {
	accessRepository repository.AccessRepository
	blackList        cache.BlackListRepository
	dbPool           *pgxpool.Pool
}

func NewAccessService(accessRepository repository.AccessRepository, blackList cache.BlackListRepository, dbPool *pgxpool.Pool) *serv {
	return &serv{accessRepository: accessRepository, blackList: blackList, dbPool: dbPool}
}
