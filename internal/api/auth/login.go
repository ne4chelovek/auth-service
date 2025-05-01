package auth

import (
	"context"
	"github.com/ne4chelovek/auth-service/internal/converter"
	desc "github.com/ne4chelovek/auth-service/pkg/auth_v1"
)

func (i *impl) Login(ctx context.Context, req *desc.LoginRequest) (*desc.LoginResponse, error) {
	refreshToken, err := i.authService.Login(ctx, converter.FromAuthDescToLogin(req.GetLogin()))
	if err != nil {
		return nil, err
	}

	return &desc.LoginResponse{
		RefreshToken: refreshToken,
	}, nil
}
