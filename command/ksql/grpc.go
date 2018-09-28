package ksql

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
	proto "github.com/confluentinc/cli/shared/ksql"
)

// GRPCClient is an implementation of Counter that talks over RPC.
type GRPCClient struct {
	client proto.KsqlClient
}

func (c *GRPCClient) List(ctx context.Context, cluster *schedv1.KSQLCluster) ([]*schedv1.KSQLCluster, error) {
	resp, err := c.client.List(ctx, &schedv1.GetKSQLClustersRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Clusters, nil
}

func (c *GRPCClient) Describe(ctx context.Context, cluster *schedv1.KSQLCluster) (*schedv1.KSQLCluster, error) {
	resp, err := c.client.Describe(ctx, &schedv1.GetKSQLClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) Delete(ctx context.Context, cluster *schedv1.KSQLCluster) error {
	_, err := c.client.Delete(ctx, &schedv1.DeleteKSQLClusterRequest{Cluster: cluster})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

func (c *GRPCClient) Create(ctx context.Context, config *schedv1.KSQLClusterConfig) (*schedv1.KSQLCluster, error) {
	resp, err := c.client.Create(ctx, &schedv1.CreateKSQLClusterRequest{Config: config})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

// GRPCServer the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl Ksql
}

func (s *GRPCServer) List(ctx context.Context, req *schedv1.GetKSQLClustersRequest) (*schedv1.GetKSQLClustersReply, error) {
	r, err := s.Impl.List(ctx, req.Cluster)
	return &schedv1.GetKSQLClustersReply{Clusters: r}, err
}

func (s *GRPCServer) Describe(ctx context.Context, req *schedv1.GetKSQLClusterRequest) (*schedv1.GetKSQLClusterReply, error) {
	r, err := s.Impl.Describe(ctx, req.Cluster)
	return &schedv1.GetKSQLClusterReply{Cluster: r}, err
}

func (s *GRPCServer) Create(ctx context.Context, req *schedv1.CreateKSQLClusterRequest) (*schedv1.CreateKSQLClusterReply, error) {
	r, err := s.Impl.Create(ctx, req.Config)
	return &schedv1.CreateKSQLClusterReply{Cluster: r}, err
}

func (s *GRPCServer) Delete(ctx context.Context, req *schedv1.DeleteKSQLClusterRequest) (*schedv1.DeleteKSQLClusterReply, error) {
	err := s.Impl.Delete(ctx, req.Cluster)
	return &schedv1.DeleteKSQLClusterReply{}, err
}

