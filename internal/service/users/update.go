package users

import (
	"context"
	"fmt"
	"github.com/ne4chelovek/auth-service/internal/model"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *serv) Update(ctx context.Context, userUp *model.UpdateUser) (*emptypb.Empty, error) {

	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	txRepo := s.usersRepository.WithTx(tx)
	txLog := s.logRepository.WithTx(tx)

	_, err = txRepo.Update(ctx, userUp)
	if err != nil {
		return nil, err
	}

	if err = txLog.Log(ctx, fmt.Sprintf("update user with id: %v", userUp.ID)); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
