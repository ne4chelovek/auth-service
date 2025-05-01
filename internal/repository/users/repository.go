package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ne4chelovek/auth-service/internal/model"
	"github.com/ne4chelovek/auth-service/internal/repository"

	sq "github.com/Masterminds/squirrel"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableUsers = "users"

	idColumn        = "id"
	nameColumn      = "name"
	emailColumn     = "email"
	passwordColumn  = "password_hash"
	roleColumn      = "role"
	createdAtColumn = "created_at"
	updatedAtColumn = "updated_at"
)

type repo struct {
	db repository.QueryRunner
}

func NewRepository(db *pgxpool.Pool) repository.UsersRepository {
	return &repo{db: db}
}

func (r *repo) WithTx(tx pgx.Tx) repository.UsersRepository {
	return &repo{db: tx}
}

func (r *repo) Create(ctx context.Context, user *model.CreateUser) (int64, error) {
	// Делам вставку в таблицу users
	builderInsert := sq.Insert(tableUsers).
		PlaceholderFormat(sq.Dollar).
		Columns(nameColumn, emailColumn, passwordColumn, roleColumn).
		Values(user.Name, user.Email, user.Password, user.Role).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		return 0, err
	}

	var userID int64
	err = r.db.QueryRow(ctx, query, args...).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *repo) Get(ctx context.Context, Id int64) (*model.User, error) {
	builderSelect := sq.Select(idColumn, nameColumn, emailColumn, roleColumn, createdAtColumn).
		From(tableUsers).
		Where(sq.Eq{idColumn: Id}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := builderSelect.ToSql()
	if err != nil {
		return nil, err
	}

	var createdAt time.Time

	user := &model.User{}
	err = r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &createdAt)
	if err != nil {
		return nil, err
	}

	user.CreatedAt = timestamppb.New(createdAt)
	if err := user.CreatedAt.CheckValid(); err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	return user, nil
}

func (r *repo) Update(ctx context.Context, userUp *model.UpdateUser) (*emptypb.Empty, error) {
	builderUpdate := sq.Update(tableUsers).
		Set(nameColumn, userUp.Name).
		Set(emailColumn, userUp.Email).
		Set(passwordColumn, userUp.Password).
		Set(updatedAtColumn, sq.Expr("NOW()")).
		Where(sq.Eq{idColumn: userUp.ID}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := builderUpdate.ToSql()
	if err != nil {
		return nil, err
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("user with ID %d not found", userUp.ID)
	}

	return &emptypb.Empty{}, nil
}

func (r *repo) Delete(ctx context.Context, Id int64) (*emptypb.Empty, error) {
	builderDelete := sq.Delete(tableUsers).
		Where(sq.Eq{idColumn: Id}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := builderDelete.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (r *repo) GetAuthInfo(ctx context.Context, name string) (*model.UserInfo, error) {
	builderSelect := sq.Select(nameColumn, passwordColumn, roleColumn).
		From(tableUsers).
		Where(sq.Eq{nameColumn: name}).
		PlaceholderFormat(sq.Dollar).
		Limit(1)

	query, args, err := builderSelect.ToSql()
	if err != nil {
		fmt.Errorf("NOT FOUND: %v", err)
		return nil, err
	}

	user := &model.UserInfo{}

	err = r.db.QueryRow(ctx, query, args...).Scan(&user.Username, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user %q not found", name)
		}
		return nil, fmt.Errorf("failed to get user auth info: %w", err)
	}

	return user, nil
}
