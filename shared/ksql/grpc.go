package ksql

import (
	"context"

	ksqlv1 "github.com/confluentinc/ccloudapis/ksql/v1"
	"github.com/confluentinc/cli/shared"
	chttp "github.com/confluentinc/ccloud-sdk-go"
)

const Name = "confluent-ksql-plugin"

var _ chttp.KSQL = (*GRPCClient)(nil)

// GRPCClient is an implementation of Counter that talks over RPC.
type GRPCClient struct {
	client KSQLClient
}

func (c *GRPCClient) List(ctx context.Context, cluster *ksqlv1.Cluster) ([]*ksqlv1.Cluster, error) {
	resp, err := c.client.List(ctx, &ksqlv1.GetClustersRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Clusters, nil
}

func (c *GRPCClient) Describe(ctx context.Context, cluster *ksqlv1.Cluster) (*ksqlv1.Cluster, error) {
	resp, err := c.client.Describe(ctx, &ksqlv1.GetClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) Delete(ctx context.Context, cluster *ksqlv1.Cluster) error {
	_, err := c.client.Delete(ctx, &ksqlv1.DeleteClusterRequest{Cluster: cluster})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

func (c *GRPCClient) Create(ctx context.Context, config *ksqlv1.ClusterConfig) (*ksqlv1.Cluster, error) {
	resp, err := c.client.Create(ctx, &ksqlv1.CreateClusterRequest{Config: config})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

var _ KSQLServer = (*GRPCServer)(nil)

// GRPCServer the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl chttp.KSQL
}

func (s *GRPCServer) List(ctx context.Context, req *ksqlv1.GetClustersRequest) (*ksqlv1.GetClustersReply, error) {
	r, err := s.Impl.List(ctx, req.Cluster)
	return &ksqlv1.GetClustersReply{Clusters: r}, err
}

func (s *GRPCServer) Describe(ctx context.Context, req *ksqlv1.GetClusterRequest) (*ksqlv1.GetClusterReply, error) {
	r, err := s.Impl.Describe(ctx, req.Cluster)
	return &ksqlv1.GetClusterReply{Cluster: r}, err
}

func (s *GRPCServer) Create(ctx context.Context, req *ksqlv1.CreateClusterRequest) (*ksqlv1.CreateClusterReply, error) {
	r, err := s.Impl.Create(ctx, req.Config)
	return &ksqlv1.CreateClusterReply{Cluster: r}, err
}

func (s *GRPCServer) Delete(ctx context.Context, req *ksqlv1.DeleteClusterRequest) (*ksqlv1.DeleteClusterReply, error) {
	err := s.Impl.Delete(ctx, req.Cluster)
	return &ksqlv1.DeleteClusterReply{}, err
}
