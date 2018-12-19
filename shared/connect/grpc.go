package connect

import (
	"context"

	connectv1 "github.com/confluentinc/ccloudapis/connect/v1"
	chttp "github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/cli/shared"
)

var _ chttp.Connect = (*GRPCClient)(nil)

// GRPCClient bridges the gap between the shared interface and the GRPC interface
type GRPCClient struct {
	client ConnectClient
}

func (c *GRPCClient) List(ctx context.Context, cluster *connectv1.Cluster) ([]*connectv1.Cluster, error) {
	resp, err := c.client.List(ctx, &connectv1.GetClustersRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Clusters, nil
}

func (c *GRPCClient) Describe(ctx context.Context, cluster *connectv1.Cluster) (*connectv1.Cluster, error) {
	resp, err := c.client.Describe(ctx, &connectv1.GetClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) DescribeS3Sink(ctx context.Context, cluster *connectv1.S3SinkCluster) (*connectv1.S3SinkCluster, error) {
	resp, err := c.client.DescribeS3Sink(ctx, &connectv1.GetS3SinkClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) CreateS3Sink(ctx context.Context, config *connectv1.S3SinkClusterConfig) (*connectv1.S3SinkCluster, error) {
	resp, err := c.client.CreateS3Sink(ctx, &connectv1.CreateS3SinkClusterRequest{Config: config})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) UpdateS3Sink(ctx context.Context, cluster *connectv1.S3SinkCluster) (*connectv1.S3SinkCluster, error) {
	resp, err := c.client.UpdateS3Sink(ctx, &connectv1.UpdateS3SinkClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) Delete(ctx context.Context, cluster *connectv1.Cluster) error {
	_, err := c.client.Delete(ctx, &connectv1.DeleteClusterRequest{Cluster: cluster})
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

func (s *GRPCServer) List(ctx context.Context, req *connectv1.GetClustersRequest) (*connectv1.GetClustersReply, error) {
	r, err := s.Impl.List(ctx, req.Cluster)
	return &connectv1.GetClustersReply{Clusters: r}, err
}

func (s *GRPCServer) Describe(ctx context.Context, req *connectv1.GetClusterRequest) (*connectv1.GetClusterReply, error) {
	r, err := s.Impl.Describe(ctx, req.Cluster)
	return &connectv1.GetClusterReply{Cluster: r}, err
}

func (s *GRPCServer) DescribeS3Sink(ctx context.Context, req *connectv1.GetS3SinkClusterRequest) (*connectv1.GetS3SinkClusterReply, error) {
	r, err := s.Impl.DescribeS3Sink(ctx, req.Cluster)
	return &connectv1.GetS3SinkClusterReply{Cluster: r}, err
}

func (s *GRPCServer) CreateS3Sink(ctx context.Context, req *connectv1.CreateS3SinkClusterRequest) (*connectv1.CreateS3SinkClusterReply, error) {
	r, err := s.Impl.CreateS3Sink(ctx, req.Config)
	return &connectv1.CreateS3SinkClusterReply{Cluster: r}, err
}

func (s *GRPCServer) UpdateS3Sink(ctx context.Context, req *connectv1.UpdateS3SinkClusterRequest) (*connectv1.UpdateS3SinkClusterReply, error) {
	r, err := s.Impl.UpdateS3Sink(ctx, req.Cluster)
	return &connectv1.UpdateS3SinkClusterReply{Cluster: r}, err
}

func (s *GRPCServer) Delete(ctx context.Context, req *connectv1.DeleteClusterRequest) (*connectv1.DeleteClusterReply, error) {
	err := s.Impl.Delete(ctx, req.Cluster)
	return &connectv1.DeleteClusterReply{}, err
}
