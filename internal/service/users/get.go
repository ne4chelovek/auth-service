package users

import (
	"context"
	"fmt"
	"github.com/ne4chelovek/auth-service/internal/model"

	"github.com/jackc/pgx/v5"
)

func (s *serv) Get(ctx context.Context, Id int64) (*model.User, error) {
	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	txRepo := s.usersRepository.WithTx(tx)
	txLog := s.logRepository.WithTx(tx)

	user, err := txRepo.Get(ctx, Id)
	if err != nil {
		return nil, err
	}

	if err := txLog.Log(ctx, fmt.Sprintf("get user with id: %v", Id)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return user, nil
}
