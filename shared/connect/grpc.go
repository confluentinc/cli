package connect

import (
	"context"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	connectv1 "github.com/confluentinc/ccloudapis/connect/v1"
	"github.com/confluentinc/cli/shared"
)

var _ chttp.Connect = (*GRPCClient)(nil)

// GRPCClient bridges the gap between the shared interface and the GRPC interface
type GRPCClient struct {
	client ConnectClient
}

func (c *GRPCClient) List(ctx context.Context, cluster *connectv1.ConnectCluster) ([]*connectv1.ConnectCluster, error) {
	resp, err := c.client.List(ctx, &connectv1.GetConnectClustersRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Clusters, nil
}

func (c *GRPCClient) Describe(ctx context.Context, cluster *connectv1.ConnectCluster) (*connectv1.ConnectCluster, error) {
	resp, err := c.client.Describe(ctx, &connectv1.GetConnectClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) DescribeS3Sink(ctx context.Context, cluster *connectv1.ConnectS3SinkCluster) (*connectv1.ConnectS3SinkCluster, error) {
	resp, err := c.client.DescribeS3Sink(ctx, &connectv1.GetConnectS3SinkClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) CreateS3Sink(ctx context.Context, config *connectv1.ConnectS3SinkClusterConfig) (*connectv1.ConnectS3SinkCluster, error) {
	resp, err := c.client.CreateS3Sink(ctx, &connectv1.CreateConnectS3SinkClusterRequest{Config: config})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) UpdateS3Sink(ctx context.Context, cluster *connectv1.ConnectS3SinkCluster) (*connectv1.ConnectS3SinkCluster, error) {
	resp, err := c.client.UpdateS3Sink(ctx, &connectv1.UpdateConnectS3SinkClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) Delete(ctx context.Context, cluster *connectv1.ConnectCluster) error {
	_, err := c.client.Delete(ctx, &connectv1.DeleteConnectClusterRequest{Cluster: cluster})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

var _ ConnectServer = (*GRPCServer)(nil)

// GRPCServer bridges the gap between the plugin implementation and the GRPC interface
type GRPCServer struct {
	Impl chttp.Connect
}

func (s *GRPCServer) List(ctx context.Context, req *connectv1.GetConnectClustersRequest) (*connectv1.GetConnectClustersReply, error) {
	r, err := s.Impl.List(ctx, req.Cluster)
	return &connectv1.GetConnectClustersReply{Clusters: r}, err
}

func (s *GRPCServer) Describe(ctx context.Context, req *connectv1.GetConnectClusterRequest) (*connectv1.GetConnectClusterReply, error) {
	r, err := s.Impl.Describe(ctx, req.Cluster)
	return &connectv1.GetConnectClusterReply{Cluster: r}, err
}

func (s *GRPCServer) DescribeS3Sink(ctx context.Context, req *connectv1.GetConnectS3SinkClusterRequest) (*connectv1.GetConnectS3SinkClusterReply, error) {
	r, err := s.Impl.DescribeS3Sink(ctx, req.Cluster)
	return &connectv1.GetConnectS3SinkClusterReply{Cluster: r}, err
}

func (s *GRPCServer) CreateS3Sink(ctx context.Context, req *connectv1.CreateConnectS3SinkClusterRequest) (*connectv1.CreateConnectS3SinkClusterReply, error) {
	r, err := s.Impl.CreateS3Sink(ctx, req.Config)
	return &connectv1.CreateConnectS3SinkClusterReply{Cluster: r}, err
}

func (s *GRPCServer) UpdateS3Sink(ctx context.Context, req *connectv1.UpdateConnectS3SinkClusterRequest) (*connectv1.UpdateConnectS3SinkClusterReply, error) {
	r, err := s.Impl.UpdateS3Sink(ctx, req.Cluster)
	return &connectv1.UpdateConnectS3SinkClusterReply{Cluster: r}, err
}

func (s *GRPCServer) Delete(ctx context.Context, req *connectv1.DeleteConnectClusterRequest) (*connectv1.DeleteConnectClusterReply, error) {
	err := s.Impl.Delete(ctx, req.Cluster)
	return &connectv1.DeleteConnectClusterReply{}, err
}
