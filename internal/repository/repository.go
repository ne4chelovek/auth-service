package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ne4chelovek/auth-service/internal/model"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

type QueryRunner interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type UsersRepository interface {
	WithTx(tx pgx.Tx) UsersRepository
	Create(ctx context.Context, user *model.CreateUser) (int64, error)
	Get(ctx context.Context, Id int64) (*model.User, error)
	Update(ctx context.Context, userUp *model.UpdateUser) (*emptypb.Empty, error)
	Delete(ctx context.Context, Id int64) (*emptypb.Empty, error)
	GetAuthInfo(ctx context.Context, name string) (*model.UserInfo, error)
}

type LogRepository interface {
	WithTx(tx pgx.Tx) LogRepository
	Log(ctx context.Context, log string) error
}

type AccessRepository interface {
	GetRoleEndpoints(ctx context.Context, endpoint string) ([]string, error)
}
