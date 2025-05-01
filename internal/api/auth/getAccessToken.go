package auth

import (
	"context"
	desc "github.com/ne4chelovek/auth-service/pkg/auth_v1"
)

func (i *impl) GetAccessToken(ctx context.Context, req *desc.GetAccessTokenRequest) (*desc.GetAccessTokenResponse, error) {
	accessToken, err := i.authService.GetAccessToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return &desc.GetAccessTokenResponse{
		AccessToken: accessToken,
	}, nil
}
