package auth

import (
	"context"
	desc "github.com/ne4chelovek/auth-service/pkg/auth_v1"
)

func (i *impl) Login(ctx context.Context, req *desc.LoginRequest) (*desc.LoginResponse, error) {
	refreshToken, err := i.authService.Login(ctx, req.Usernames, req.Password)
	if err != nil {
		return nil, err
	}

	return &desc.LoginResponse{
		RefreshToken: refreshToken,
	}, nil
}
