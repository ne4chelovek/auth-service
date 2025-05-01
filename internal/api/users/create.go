package users

import (
	"context"
	"github.com/ne4chelovek/auth-service/internal/converter"
	desc "github.com/ne4chelovek/auth-service/pkg/users_v1"
)

func (i *impl) Create(ctx context.Context, newUser *desc.CreateRequest) (*desc.CreateResponse, error) {
	id, err := i.usersService.Create(ctx, converter.FromDescCreateToUser(newUser.User))
	if err != nil {
		return nil, err
	}

	return &desc.CreateResponse{
		Id: id,
	}, nil
}
