package users

import (
	"github.com/ne4chelovek/auth-service/internal/repository"
	"github.com/ne4chelovek/auth-service/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type serv struct {
	usersRepository repository.UsersRepository
	dbPool          *pgxpool.Pool
	logRepository   repository.LogRepository
}

var _ service.UsersService = (*serv)(nil)

func NewUsersService(usersRepository repository.UsersRepository, logRepository repository.LogRepository, dbPool *pgxpool.Pool) service.UsersService {
	return &serv{
		usersRepository: usersRepository,
		logRepository:   logRepository,
		dbPool:          dbPool,
	}
}
