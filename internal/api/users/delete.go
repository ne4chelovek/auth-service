package users

import (
	"context"
	desc "github.com/ne4chelovek/auth-service/pkg/users_v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *impl) Delete(ctx context.Context, Id *desc.DeleteRequest) (*emptypb.Empty, error) {
	_, err := i.usersService.Delete(ctx, Id.Id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
