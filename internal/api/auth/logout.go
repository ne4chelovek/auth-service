package auth

import (
	"context"
	desc "github.com/ne4chelovek/auth-service/pkg/auth_v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *impl) Logout(ctx context.Context, req *desc.LogoutRequest) (*emptypb.Empty, error) {
	if err := i.authService.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
