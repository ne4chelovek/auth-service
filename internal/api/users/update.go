package users

import (
	"context"
	"github.com/ne4chelovek/auth-service/internal/converter"
	desc "github.com/ne4chelovek/auth-service/pkg/users_v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *impl) Update(ctx context.Context, upUser *desc.UpdateRequest) (*emptypb.Empty, error) {
	_, err := i.usersService.Update(ctx, converter.FromDescUpdateToAuth(upUser.User))
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
