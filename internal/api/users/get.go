package users

import (
	"context"
	"github.com/ne4chelovek/auth-service/internal/converter"
	desc "github.com/ne4chelovek/auth-service/pkg/users_v1"
)

func (i *impl) Get(ctx context.Context, Id *desc.GetRequest) (*desc.GetResponse, error) {
	info, err := i.usersService.Get(ctx, Id.Id)
	if err != nil {
		return nil, err
	}

	return &desc.GetResponse{
		User: converter.FromUserToDesc(info),
	}, nil
}
