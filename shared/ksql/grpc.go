package ksql

import (
	"context"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	ksqlv1 "github.com/confluentinc/ccloudapis/ksql/v1"
	"github.com/confluentinc/cli/shared"
)

var _ chttp.KSQL = (*GRPCClient)(nil)

// GRPCClient is an implementation of Counter that talks over RPC.
type GRPCClient struct {
	client KSQLClient
}

func (c *GRPCClient) List(ctx context.Context, cluster *ksqlv1.KSQLCluster) ([]*ksqlv1.KSQLCluster, error) {
	resp, err := c.client.List(ctx, &ksqlv1.GetKSQLClustersRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Clusters, nil
}

func (c *GRPCClient) Describe(ctx context.Context, cluster *ksqlv1.KSQLCluster) (*ksqlv1.KSQLCluster, error) {
	resp, err := c.client.Describe(ctx, &ksqlv1.GetKSQLClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) Delete(ctx context.Context, cluster *ksqlv1.KSQLCluster) error {
	_, err := c.client.Delete(ctx, &ksqlv1.DeleteKSQLClusterRequest{Cluster: cluster})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

func (c *GRPCClient) Create(ctx context.Context, config *ksqlv1.KSQLClusterConfig) (*ksqlv1.KSQLCluster, error) {
	resp, err := c.client.Create(ctx, &ksqlv1.CreateKSQLClusterRequest{Config: config})
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

func (s *GRPCServer) List(ctx context.Context, req *ksqlv1.GetKSQLClustersRequest) (*ksqlv1.GetKSQLClustersReply, error) {
	r, err := s.Impl.List(ctx, req.Cluster)
	return &ksqlv1.GetKSQLClustersReply{Clusters: r}, err
}

func (s *GRPCServer) Describe(ctx context.Context, req *ksqlv1.GetKSQLClusterRequest) (*ksqlv1.GetKSQLClusterReply, error) {
	r, err := s.Impl.Describe(ctx, req.Cluster)
	return &ksqlv1.GetKSQLClusterReply{Cluster: r}, err
}

func (s *GRPCServer) Create(ctx context.Context, req *ksqlv1.CreateKSQLClusterRequest) (*ksqlv1.CreateKSQLClusterReply, error) {
	r, err := s.Impl.Create(ctx, req.Config)
	return &ksqlv1.CreateKSQLClusterReply{Cluster: r}, err
}

func (s *GRPCServer) Delete(ctx context.Context, req *ksqlv1.DeleteKSQLClusterRequest) (*ksqlv1.DeleteKSQLClusterReply, error) {
	err := s.Impl.Delete(ctx, req.Cluster)
	return &ksqlv1.DeleteKSQLClusterReply{}, err
}
