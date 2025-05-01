package users

import (
	"github.com/ne4chelovek/auth-service/internal/service"
	desc "github.com/ne4chelovek/auth-service/pkg/users_v1"
)

type impl struct {
	desc.UnimplementedUsersV1Server
	usersService service.UsersService
}

func NewUsersImplementation(usersService service.UsersService) *impl {
	return &impl{
		usersService: usersService,
	}
}
