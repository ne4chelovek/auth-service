package access

import (
	"github.com/ne4chelovek/auth-service/internal/service"
	desc "github.com/ne4chelovek/auth-service/pkg/access_v1"
)

type impl struct {
	desc.UnimplementedAccessV1Server
	accessService service.AccessService
}

func NewAccessTokenImplementation(accessService service.AccessService) *impl {
	return &impl{
		accessService: accessService,
	}
}
