package users

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *serv) Delete(ctx context.Context, Id int64) (*emptypb.Empty, error) {

	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	txRepo := s.usersRepository.WithTx(tx)
	logTxrepo := s.logRepository.WithTx(tx)

	_, err = txRepo.Delete(ctx, Id)
	if err != nil {
		return nil, err
	}

	if err := logTxrepo.Log(ctx, fmt.Sprintf("delete user with id: %v", Id)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
