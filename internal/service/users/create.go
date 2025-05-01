package users

import (
	"fmt"
	"github.com/ne4chelovek/auth-service/internal/model"
	"golang.org/x/crypto/bcrypt"

	"context"

	"github.com/jackc/pgx/v5"
)

// TODO: Сделать сопоставление полученных данных с шаблоном, поможет пакет regexp
func (s *serv) Create(ctx context.Context, user *model.CreateUser) (int64, error) {
	if user.Password != user.PasswordConfirm {
		return 0, fmt.Errorf("password does not match")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("Слабый пароль  %v", err)
	}

	user.Password = string(hashedPassword)

	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	txRepo := s.usersRepository.WithTx(tx)
	logTxRepo := s.logRepository.WithTx(tx)

	userID, err := txRepo.Create(ctx, user)
	if err != nil {
		return 0, err
	}

	if err := logTxRepo.Log(ctx, fmt.Sprintf("create user with id: %v", userID)); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return userID, nil
}
