package kafka

import (
	"context"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
)

// GRPCClient is an implementation of Counter that talks over RPC.
type GRPCClient struct {
	client KafkaClient
}

// CreateAPIKey creates a new API key for accessing a Kafka Cluster
func (c *GRPCClient) CreateAPIKey(ctx context.Context, apiKey *schedv1.ApiKey) (*schedv1.ApiKey, error) {
	r, err := c.client.CreateAPIKey(ctx, &schedv1.CreateApiKeyRequest{ApiKey: apiKey})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return r.ApiKey, err
}

// List returns a list of Kafka Cluster available to the authenticated user
func (c *GRPCClient) List(ctx context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error) {
	resp, err := c.client.List(ctx, &schedv1.GetKafkaClustersRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Clusters, nil
}

// Describe provides detailed information about a Kafka Cluster
func (c *GRPCClient) Describe(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
	resp, err := c.client.Describe(ctx, &schedv1.GetKafkaClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

// Create generates a new Kafka Cluster
func (c *GRPCClient) Create(ctx context.Context, config *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, error) {
	resp, err := c.client.Create(ctx, &schedv1.CreateKafkaClusterRequest{Config: config})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

// Delete removes a Kafka Cluster
func (c *GRPCClient) Delete(ctx context.Context, cluster *schedv1.KafkaCluster) error {
	_, err := c.client.Delete(ctx, &schedv1.DeleteKafkaClusterRequest{Cluster: cluster})
	return shared.ConvertGRPCError(err)
}

// ListTopics lists all non-internal topics in the current Kafka cluster context
func (c *GRPCClient) ListTopics(ctx context.Context, cluster *schedv1.KafkaCluster) ([]string, error) {
	r, err := c.client.ListTopics(ctx, &KafkaTopicRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return r.Topics, shared.ConvertGRPCError(err)
}

// DescribeTopic returns details for a Kafka Topic in the current Kafka Cluster context
func (c *GRPCClient) DescribeTopic(ctx context.Context, cluster *schedv1.KafkaCluster, topic *Topic) (*KafkaTopicDescription, error) {
	r, err := c.client.DescribeTopic(ctx, &KafkaTopicRequest{Cluster: cluster, Topic: topic})
	return r, shared.ConvertGRPCError(err)
}

// CreateTopic creates a new Kafka Topic in the current Kafka Cluster context
func (c *GRPCClient) CreateTopic(ctx context.Context, cluster *schedv1.KafkaCluster, topic *Topic) (*KafkaAPIResponse, error) {
	r, err := c.client.CreateTopic(ctx, &KafkaTopicRequest{Cluster: cluster, Topic: topic})
	return r, shared.ConvertGRPCError(err)
}

// DeleteTopic a Kafka Topic in the current Kafka Cluster context
func (c *GRPCClient) DeleteTopic(ctx context.Context, cluster *schedv1.KafkaCluster, topic *Topic) (*KafkaAPIResponse, error) {
	r, err := c.client.DeleteTopic(ctx, &KafkaTopicRequest{Cluster: cluster, Topic: topic})
	return r, shared.ConvertGRPCError(err)
}

// UpdateTopic updates any existing Topic's configuration in the current Kafka Cluster context
func (c *GRPCClient) UpdateTopic(ctx context.Context, cluster *schedv1.KafkaCluster, topic *Topic) (*KafkaAPIResponse, error) {
	r, err := c.client.UpdateTopic(ctx, &KafkaTopicRequest{Cluster: cluster, Topic: topic})
	return r, shared.ConvertGRPCError(err)
}

// ListACL lists all ACLs for a given principal or resource
func (c *GRPCClient) ListACL(ctx context.Context, conf *ACLFilter) (*KafkaAPIACLFilterReply, error) {
	r, err := c.client.ListACL(ctx, conf)
	return r, shared.ConvertGRPCError(err)
}

// CreateACL registers a new ACL with the currently Kafka Cluster context
func (c *GRPCClient) CreateACL(ctx context.Context, conf *ACLSpec) (*KafkaAPIResponse, error) {
	r, err := c.client.CreateACL(ctx, conf)
	return r, shared.ConvertGRPCError(err)
}

// DeleteACL removes an ACL with the currently Kafka Cluster context
func (c *GRPCClient) DeleteACL(ctx context.Context, conf *ACLFilter) (*KafkaAPIResponse, error) {
	r, err := c.client.DeleteACL(ctx, conf)
	return r, shared.ConvertGRPCError(err)
}

// GRPCServer the GPRClient talks to. Plugin authors implement this if they're using Go.
type GRPCServer struct {
	Impl Kafka
}

// CreateAPIKey creates a new API key for accessing a kafka cluster
func (s *GRPCServer) CreateAPIKey(ctx context.Context, req *schedv1.CreateApiKeyRequest) (*schedv1.CreateApiKeyReply, error) {
	r, err := s.Impl.CreateAPIKey(ctx, req.ApiKey)
	return &schedv1.CreateApiKeyReply{ApiKey: r}, err
}

// List returns a list of Kafka Cluster available to the authenticated user
func (s *GRPCServer) List(ctx context.Context, req *schedv1.GetKafkaClustersRequest) (*schedv1.GetKafkaClustersReply, error) {
	r, err := s.Impl.List(ctx, req.Cluster)
	return &schedv1.GetKafkaClustersReply{Clusters: r}, err
}

// Describe provides detailed information about a Kafka Cluster
func (s *GRPCServer) Describe(ctx context.Context, req *schedv1.GetKafkaClusterRequest) (*schedv1.GetKafkaClusterReply, error) {
	r, err := s.Impl.Describe(ctx, req.Cluster)
	return &schedv1.GetKafkaClusterReply{Cluster: r}, err
}

// Create generates a new Kafka Cluster
func (s *GRPCServer) Create(ctx context.Context, req *schedv1.CreateKafkaClusterRequest) (*schedv1.CreateKafkaClusterReply, error) {
	r, err := s.Impl.Create(ctx, req.Config)
	return &schedv1.CreateKafkaClusterReply{Cluster: r}, err
}

// Delete removes a Kafka Cluster
func (s *GRPCServer) Delete(ctx context.Context, req *schedv1.DeleteKafkaClusterRequest) (*schedv1.DeleteKafkaClusterReply, error) {
	err := s.Impl.Delete(ctx, req.Cluster)
	return &schedv1.DeleteKafkaClusterReply{}, err
}

// ListTopics lists all non-internal topics in the current Kafka Cluster context
func (s *GRPCServer) ListTopics(ctx context.Context, req *KafkaTopicRequest) (*ListKafkaTopicReply, error) {
	topics, err := s.Impl.ListTopics(ctx, req.Cluster)
	return &ListKafkaTopicReply{Topics: topics}, err
}

// DescribeTopic returns details for a Kafka Topic in the current Kafka Cluster context
func (s *GRPCServer) DescribeTopic(ctx context.Context, req *KafkaTopicRequest) (*KafkaTopicDescription, error) {
	return s.Impl.DescribeTopic(ctx, req.Cluster, req.Topic)
}

// CreateTopic creates a new Kafka Topic in the current Kafka Cluster context
func (s *GRPCServer) CreateTopic(ctx context.Context, req *KafkaTopicRequest) (*KafkaAPIResponse, error) {
	return s.Impl.CreateTopic(ctx, req.Cluster, req.Topic)
}

// DeleteTopic deletes a Kafka Topic in the current Kafka Cluster context
func (s *GRPCServer) DeleteTopic(ctx context.Context, req *KafkaTopicRequest) (*KafkaAPIResponse, error) {
	return s.Impl.DeleteTopic(ctx, req.Cluster, req.Topic)
}

// UpdateTopic updates any existing Topic's configuration in the current Kafka Cluster context
func (s *GRPCServer) UpdateTopic(ctx context.Context, req *KafkaTopicRequest) (*KafkaAPIResponse, error) {
	return s.Impl.UpdateTopic(ctx, req.Cluster, req.Topic)
}

// ListACL lists all ACLs for a given principal or resource
func (s *GRPCServer) ListACL(ctx context.Context, conf *ACLFilter) (*KafkaAPIACLFilterReply, error) {
	return s.Impl.ListACL(ctx, conf)
}

// CreateACL registers a new ACL with the currently Kafka Cluster context
func (s *GRPCServer) CreateACL(ctx context.Context, conf *ACLSpec) (*KafkaAPIResponse, error) {
	return s.Impl.CreateACL(ctx, conf)
}

// DeleteACL removes an ACL with the currently Kafka Cluster context
func (s *GRPCServer) DeleteACL(ctx context.Context, conf *ACLFilter) (*KafkaAPIResponse, error) {
	return s.Impl.DeleteACL(ctx, conf)
}
