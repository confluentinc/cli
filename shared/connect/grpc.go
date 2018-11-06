package connect

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
)

// GRPCClient extends GRPC ConnectClient with plugin specific interfaces
type GRPCClient struct {
	client ConnectClient
}

// List returns a List of connect clusters
func (c *GRPCClient) List(ctx context.Context, cluster *schedv1.ConnectCluster) ([]*schedv1.ConnectCluster, error) {
	resp, err := c.client.List(ctx, &schedv1.GetConnectClustersRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Clusters, nil
}

// DescribeS3Sink returns the configuration for an s3 sink
func (c *GRPCClient) DescribeS3Sink(ctx context.Context, cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, error) {
	resp, err := c.client.DescribeS3Sink(ctx, &schedv1.GetConnectS3SinkClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

// Describe returns the configuration for connect
func (c *GRPCClient) Describe(ctx context.Context, cluster *schedv1.ConnectCluster) (*schedv1.ConnectCluster, error) {
	resp, err := c.client.Describe(ctx, &schedv1.GetConnectClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

// CreateS3Sink registers a new S3SinkConnector with connect
func (c *GRPCClient) CreateS3Sink(ctx context.Context, config *ConnectS3SinkClusterConfig) (*schedv1.ConnectS3SinkCluster, error) {
	resp, err := c.client.CreateS3Sink(ctx, &CreateConnectS3SinkClusterRequest{Config: config})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

// UpdateS3Sink updates a S3SinkConnector's configuration
func (c *GRPCClient) UpdateS3Sink(ctx context.Context, cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, error) {
	resp, err := c.client.UpdateS3Sink(ctx, &schedv1.UpdateConnectS3SinkClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

// Delete removes a Connect Cluster
func (c *GRPCClient) Delete(ctx context.Context, cluster *schedv1.ConnectCluster) error {
	_, err := c.client.Delete(ctx, &schedv1.DeleteConnectClusterRequest{Cluster: cluster})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

// GRPCServer the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl Connect
}

// List returns a List of connect clusters
func (s *GRPCServer) List(ctx context.Context, req *schedv1.GetConnectClustersRequest) (*schedv1.GetConnectClustersReply, error) {
	r, err := s.Impl.List(ctx, req.Cluster)
	return &schedv1.GetConnectClustersReply{Clusters: r}, err
}

// Describe returns the configuration for connect
func (s *GRPCServer) Describe(ctx context.Context, req *schedv1.GetConnectClusterRequest) (*schedv1.GetConnectClusterReply, error) {
	r, err := s.Impl.Describe(ctx, req.Cluster)
	return &schedv1.GetConnectClusterReply{Cluster: r}, err
}

// DescribeS3Sink returns the configuration for an s3 sink
func (s *GRPCServer) DescribeS3Sink(ctx context.Context, req *schedv1.GetConnectS3SinkClusterRequest) (*schedv1.GetConnectS3SinkClusterReply, error) {
	r, err := s.Impl.DescribeS3Sink(ctx, req.Cluster)
	return &schedv1.GetConnectS3SinkClusterReply{Cluster: r}, err
}

// CreateS3Sink registers a new S3SinkConnector with connect
func (s *GRPCServer) CreateS3Sink(ctx context.Context, req *CreateConnectS3SinkClusterRequest) (*CreateConnectS3SinkClusterReply, error) {
	r, err := s.Impl.CreateS3Sink(ctx, req.Config)
	return &CreateConnectS3SinkClusterReply{Cluster: r}, err
}

// UpdateS3Sink updates a S3SinkConnector's configuration
func (s *GRPCServer) UpdateS3Sink(ctx context.Context, req *schedv1.UpdateConnectS3SinkClusterRequest) (*schedv1.UpdateConnectS3SinkClusterReply, error) {
	r, err := s.Impl.UpdateS3Sink(ctx, req.Cluster)
	return &schedv1.UpdateConnectS3SinkClusterReply{Cluster: r}, err
}

// Delete removes a Connect Cluster
func (s *GRPCServer) Delete(ctx context.Context, req *schedv1.DeleteConnectClusterRequest) (*schedv1.DeleteConnectClusterReply, error) {
	err := s.Impl.Delete(ctx, req.Cluster)
	return &schedv1.DeleteConnectClusterReply{}, err
}
