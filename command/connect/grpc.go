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

func (c *GRPCClient) List(ctx context.Context, cluster *schedv1.ConnectCluster) ([]*schedv1.ConnectCluster, error) {
	resp, err := c.client.List(ctx, &schedv1.GetConnectClustersRequest{Cluster: cluster})
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

func (c *GRPCClient) CreateS3Sink(ctx context.Context, config *schedv1.ConnectS3SinkClusterConfig) (*schedv1.ConnectS3SinkCluster, error) {
	resp, err := c.client.CreateS3Sink(ctx, &schedv1.CreateConnectS3SinkClusterRequest{Config: config})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) UpdateS3Sink(ctx context.Context, cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, error) {
	resp, err := c.client.UpdateS3Sink(ctx, &schedv1.UpdateConnectS3SinkClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) Delete(ctx context.Context, cluster *schedv1.ConnectCluster) error {
	_, err := c.client.Delete(ctx, &schedv1.DeleteConnectClusterRequest{Cluster: cluster})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

// The gRPC server the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl Connect
}

func (s *GRPCServer) List(ctx context.Context, req *schedv1.GetConnectClustersRequest) (*schedv1.GetConnectClustersReply, error) {
	r, err := s.Impl.List(ctx, req.Cluster)
	return &schedv1.GetConnectClustersReply{Clusters: r}, err
}

func (s *GRPCServer) Describe(ctx context.Context, req *schedv1.GetConnectClusterRequest) (*schedv1.GetConnectClusterReply, error) {
	r, err := s.Impl.Describe(ctx, req.Cluster)
	return &schedv1.GetConnectClusterReply{Cluster: r}, err
}

func (s *GRPCServer) CreateS3Sink(ctx context.Context, req *schedv1.CreateConnectS3SinkClusterRequest) (*schedv1.CreateConnectS3SinkClusterReply, error) {
	r, err := s.Impl.CreateS3Sink(ctx, req.Config)
	return &schedv1.CreateConnectS3SinkClusterReply{Cluster: r}, err
}

func (s *GRPCServer) UpdateS3Sink(ctx context.Context, req *schedv1.UpdateConnectS3SinkClusterRequest) (*schedv1.UpdateConnectS3SinkClusterReply, error) {
	r, err := s.Impl.UpdateS3Sink(ctx, req.Cluster)
	return &schedv1.UpdateConnectS3SinkClusterReply{Cluster: r}, err
}

func (s *GRPCServer) Delete(ctx context.Context, req *schedv1.DeleteConnectClusterRequest) (*schedv1.DeleteConnectClusterReply, error) {
	err := s.Impl.Delete(ctx, req.Cluster)
	return &schedv1.DeleteConnectClusterReply{}, err
}
