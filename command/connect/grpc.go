package connect

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
	proto "github.com/confluentinc/cli/shared/connect"
)

// GRPCClient is an implementation of Counter that talks over RPC.
type GRPCClient struct {
	client proto.ConnectClient
}

func (c *GRPCClient) List(ctx context.Context) ([]*schedv1.ConnectCluster, error) {
	resp, err := c.client.List(ctx, &schedv1.GetConnectClustersRequest{})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Clusters, nil
}

func (c *GRPCClient) Describe(ctx context.Context, cluster *schedv1.ConnectCluster) (*schedv1.ConnectCluster, error) {
	resp, err := c.client.Describe(ctx, &schedv1.GetConnectClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

// The gRPC server the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl Connect
}

func (s *GRPCServer) List(ctx context.Context, req *schedv1.GetConnectClustersRequest) (*schedv1.GetConnectClustersReply, error) {
	r, err := s.Impl.List(ctx)
	return &schedv1.GetConnectClustersReply{Clusters: r}, err
}

func (s *GRPCServer) Describe(ctx context.Context, req *schedv1.GetConnectClusterRequest) (*schedv1.GetConnectClusterReply, error) {
	r, err := s.Impl.Describe(ctx, req.Cluster)
	return &schedv1.GetConnectClusterReply{Cluster: r}, err
}
