package kafka

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
	proto "github.com/confluentinc/cli/shared/kafka"
)

// GRPCClient is an implementation of Counter that talks over RPC.
type GRPCClient struct {
	client proto.KafkaClient
}

func (c *GRPCClient) List(ctx context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error) {
	resp, err := c.client.List(ctx, &schedv1.GetKafkaClustersRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Clusters, nil
}

func (c *GRPCClient) Describe(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
	resp, err := c.client.Describe(ctx, &schedv1.GetKafkaClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) Create(ctx context.Context, config *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, error) {
	resp, err := c.client.Create(ctx, &schedv1.CreateKafkaClusterRequest{Config: config})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) Delete(ctx context.Context, cluster *schedv1.KafkaCluster) error {
	_, err := c.client.Delete(ctx, &schedv1.DeleteKafkaClusterRequest{Cluster: cluster})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

// GRPCServer the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl Kafka
}

func (s *GRPCServer) List(ctx context.Context, req *schedv1.GetKafkaClustersRequest) (*schedv1.GetKafkaClustersReply, error) {
	r, err := s.Impl.List(ctx, req.Cluster)
	return &schedv1.GetKafkaClustersReply{Clusters: r}, err
}

func (s *GRPCServer) Describe(ctx context.Context, req *schedv1.GetKafkaClusterRequest) (*schedv1.GetKafkaClusterReply, error) {
	r, err := s.Impl.Describe(ctx, req.Cluster)
	return &schedv1.GetKafkaClusterReply{Cluster: r}, err
}

func (s *GRPCServer) Create(ctx context.Context, req *schedv1.CreateKafkaClusterRequest) (*schedv1.CreateKafkaClusterReply, error) {
	r, err := s.Impl.Create(ctx, req.Config)
	return &schedv1.CreateKafkaClusterReply{Cluster: r}, err
}

func (s *GRPCServer) Delete(ctx context.Context, req *schedv1.DeleteKafkaClusterRequest) (*schedv1.DeleteKafkaClusterReply, error) {
	err := s.Impl.Delete(ctx, req.Cluster)
	return &schedv1.DeleteKafkaClusterReply{}, err
}
