package log

import (
	"context"
	"github.com/ne4chelovek/auth-service/internal/repository"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableLog = "transaction_log"
	columLog = "log"
)

type repo struct {
	db repository.QueryRunner
}

func NewLogRepository(db *pgxpool.Pool) repository.LogRepository {
	return &repo{db: db}
}

func (r *repo) WithTx(tx pgx.Tx) repository.LogRepository {
	return &repo{db: tx}
}

func (r *repo) Log(ctx context.Context, log string) error {
	builderInsert := sq.Insert(tableLog).
		PlaceholderFormat(sq.Dollar).
		Columns(columLog).
		Values(log).
		Suffix("RETURNING id")

	q, args, err := builderInsert.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}

	return nil
}
