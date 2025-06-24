package users

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/ne4chelovek/auth-service/internal/model"
)

func (s *serv) Get(ctx context.Context, Id int64) (*model.User, error) {
	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	user, err := s.usersRepository.WithTx(tx).Get(ctx, Id)
	if err != nil {
		return nil, err
	}

	if err := s.logRepository.WithTx(tx).Log(ctx, fmt.Sprintf("get user with id: %v", Id)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return user, nil
}
