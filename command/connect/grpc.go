package connect

import (
	"context"

	proto "github.com/confluentinc/cli/shared/connect"
)

// GRPCClient is an implementation of Counter that talks over RPC.
type GRPCClient struct {
	client proto.ConnectClient
}

func (c *GRPCClient) List(ctx context.Context) (connectors []*proto.Connector, err error) {
	resp, err := c.client.List(ctx, &proto.ListRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Clusters, nil
}

// The gRPC server the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl Connect
}

func (s *GRPCServer) List(ctx context.Context, req *proto.ListRequest) (resp *proto.ListResponse, err error) {
	r, err := s.Impl.List(ctx)
	return &proto.ListResponse{Clusters: r}, err
}
