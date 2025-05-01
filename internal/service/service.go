package service

import (
	"context"
	"github.com/ne4chelovek/auth-service/internal/model"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UsersService interface {
	Create(ctx context.Context, user *model.CreateUser) (int64, error)
	Get(ctx context.Context, Id int64) (*model.User, error)
	Update(ctx context.Context, userUp *model.UpdateUser) (*emptypb.Empty, error)
	Delete(ctx context.Context, Id int64) (*emptypb.Empty, error)
}

type AuthService interface {
	Login(ctx context.Context, login *model.UserCreds) (string, error)
	GetRefreshToken(ctx context.Context, oldRefreshToken string) (string, error)
	GetAccessToken(ctx context.Context, accessToken string) (string, error)
}

type AccessService interface {
	Check(ctx context.Context, endpoint string) (bool, error)
}
