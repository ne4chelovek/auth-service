package auth

import (
	"github.com/ne4chelovek/auth-service/internal/service"
	desc "github.com/ne4chelovek/auth-service/pkg/auth_v1"
)

type impl struct {
	desc.UnimplementedAuthV1Server
	authService service.AuthService
}

func NewAuthImplementation(authService service.AuthService) *impl {
	return &impl{
		authService: authService,
	}
}
