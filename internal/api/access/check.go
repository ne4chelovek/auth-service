package access

import (
	"context"
	desc "github.com/ne4chelovek/auth-service/pkg/access_v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *impl) Check(ctx context.Context, req *desc.CheckRequest) (*emptypb.Empty, error) {
	_, err := i.accessService.Check(ctx, req.GetEndpointAddress())
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
