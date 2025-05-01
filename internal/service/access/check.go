package access

import (
	"context"
	"fmt"
	"github.com/ne4chelovek/auth-service/internal/utils"
	"google.golang.org/grpc/metadata"
	"strings"
)

const (
	authMetadataHeader = "authorization"
	authPrefix         = "Bearer "
)

func (s *serv) Check(ctx context.Context, endpoint string) (bool, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false, fmt.Errorf("fail to get metadata")
	}

	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return false, fmt.Errorf("fail to get authorization header")
	}
	if !strings.HasPrefix(authHeader[0], authPrefix) {
		return false, fmt.Errorf("fail to check authorization")
	}

	accessToken := strings.TrimPrefix(authHeader[0], authPrefix)

	claims, err := utils.VerifyToken(accessToken, []byte(accessTokenSecretKeyName))
	if err != nil {
		return false, fmt.Errorf("fail to verify access token")
	}
	allowedRoles, err := s.accessRepository.GetRoleEndpoints(ctx, endpoint)
	if err != nil {
		return false, fmt.Errorf("fail to get role endpoints")
	}

	for _, role := range allowedRoles {
		if role == claims.Role {
			return true, nil
		}
	}

	return false, fmt.Errorf("fail to check accessible role")
}
