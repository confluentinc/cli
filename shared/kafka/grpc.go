package kafka

import (
	"context"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/cli/shared"
)

var _ chttp.Kafka = (*GRPCClient)(nil)

type GRPCClient struct {
	client KafkaClient
}

// GRPCClient bridges the gap between the shared interface and the GRPC interface
func (c *GRPCClient) List(ctx context.Context, cluster *kafkav1.Cluster) ([]*kafkav1.Cluster, error) {
	resp, err := c.client.List(ctx, &kafkav1.GetClustersRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Clusters, nil
}

func (c *GRPCClient) Describe(ctx context.Context, cluster *kafkav1.Cluster) (*kafkav1.Cluster, error) {
	resp, err := c.client.Describe(ctx, &kafkav1.GetClusterRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) Create(ctx context.Context, config *kafkav1.ClusterConfig) (*kafkav1.Cluster, error) {
	resp, err := c.client.Create(ctx, &kafkav1.CreateClusterRequest{Config: config})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return resp.Cluster, nil
}

func (c *GRPCClient) Delete(ctx context.Context, cluster *kafkav1.Cluster) error {
	_, err := c.client.Delete(ctx, &kafkav1.DeleteClusterRequest{Cluster: cluster})
	if err != nil {
		return shared.ConvertGRPCError(err)
	}
	return nil
}

// ListTopics lists all non-internal topics in the current Kafka cluster context
func (c *GRPCClient) ListTopics(ctx context.Context, cluster *kafkav1.Cluster) ([]*kafkav1.TopicDescription, error) {
	r, err := c.client.ListTopics(ctx, &kafkav1.ListTopicRequest{Cluster: cluster})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return r.Topics, nil
}

// DescribeTopic returns details for a Kafka Topic in the current Kafka Cluster context
func (c *GRPCClient) DescribeTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) (*kafkav1.TopicDescription, error) {
	r, err := c.client.DescribeTopic(ctx, &kafkav1.DescribeTopicRequest{Cluster: cluster, Topic: topic})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return r.Topic, nil
}

// CreateTopic creates a new Kafka Topic in the current Kafka Cluster context
func (c *GRPCClient) CreateTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) error {
	_, err := c.client.CreateTopic(ctx, &kafkav1.CreateTopicRequest{Cluster: cluster, Topic: topic})
	return shared.ConvertGRPCError(err)
}

// DeleteTopic a Kafka Topic in the current Kafka Cluster context
func (c *GRPCClient) DeleteTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) error {
	_, err := c.client.DeleteTopic(ctx, &kafkav1.DeleteTopicRequest{Cluster: cluster, Topic: topic})
	return shared.ConvertGRPCError(err)
}

// ListTopicConfig is not currently updates any existing Topic's configuration in the current Kafka Cluster context
func (c *GRPCClient) ListTopicConfig(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) (*kafkav1.TopicConfig, error) {
	return nil, shared.ErrNotImplemented
}

// UpdateTopic updates any existing Topic's configuration in the current Kafka Cluster context
func (c *GRPCClient) UpdateTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) error {
	_, err := c.client.UpdateTopic(ctx, &kafkav1.UpdateTopicRequest{Cluster: cluster, Topic: topic})
	return shared.ConvertGRPCError(err)
}

// ListACL lists all ACLs for a given principal or resource
func (c *GRPCClient) ListACL(ctx context.Context, cluster *kafkav1.Cluster, filter *kafkav1.ACLFilter) ([]*kafkav1.ACLBinding, error) {
	r, err := c.client.ListACL(ctx, &kafkav1.ListACLRequest{Cluster: cluster, Filter: filter})
	if err != nil {
		return nil, shared.ConvertGRPCError(err)
	}
	return r.Results, nil
}

// CreateACL registers a new ACL with the currently Kafka Cluster context
func (c *GRPCClient) CreateACL(ctx context.Context, cluster *kafkav1.Cluster, binding []*kafkav1.ACLBinding) error {
	_, err := c.client.CreateACL(ctx, &kafkav1.CreateACLRequest{Cluster: cluster, AclBindings: binding})
	return shared.ConvertGRPCError(err)
}

// DeleteACL removes an ACL with the currently Kafka Cluster context
func (c *GRPCClient) DeleteACL(ctx context.Context, cluster *kafkav1.Cluster, filter *kafkav1.ACLFilter) error {
	_, err := c.client.DeleteACL(ctx, &kafkav1.DeleteACLRequest{Cluster: cluster, Filter: filter})
	return shared.ConvertGRPCError(err)
}

var _ KafkaServer = (*GRPCServer)(nil)

// GRPCServer bridges the gap between the plugin implementation and the GRPC interface
type GRPCServer struct {
	Impl chttp.Kafka
}

// CreateAPIKey creates a new API key for accessing a kafka cluster
func (s *GRPCServer) CreateAPIKey(ctx context.Context, req *authv1.CreateAPIKeyRequest) (*authv1.CreateAPIKeyReply, error) {
	//r, err := s.Impl.CreateAPIKey(ctx, req.ApiKey)
	return nil, shared.ErrNotImplemented
}

// List returns a list of Kafka Cluster available to the authenticated user
func (s *GRPCServer) List(ctx context.Context, req *kafkav1.GetClustersRequest) (*kafkav1.GetClustersReply, error) {
	r, err := s.Impl.List(ctx, req.Cluster)
	return &kafkav1.GetClustersReply{Clusters: r}, err
}

// Describe provides detailed information about a Kafka Cluster
func (s *GRPCServer) Describe(ctx context.Context, req *kafkav1.GetClusterRequest) (*kafkav1.GetClusterReply, error) {
	r, err := s.Impl.Describe(ctx, req.Cluster)
	return &kafkav1.GetClusterReply{Cluster: r}, err
}

// Create generates a new Kafka Cluster
func (s *GRPCServer) Create(ctx context.Context, req *kafkav1.CreateClusterRequest) (*kafkav1.CreateClusterReply, error) {
	r, err := s.Impl.Create(ctx, req.Config)
	return &kafkav1.CreateClusterReply{Cluster: r}, err
}

// Delete removes a Kafka Cluster
func (s *GRPCServer) Delete(ctx context.Context, req *kafkav1.DeleteClusterRequest) (*kafkav1.DeleteClusterReply, error) {
	err := s.Impl.Delete(ctx, req.Cluster)
	return &kafkav1.DeleteClusterReply{}, err
}

// ListTopics lists all non-internal topics in the current Kafka Cluster context
func (s *GRPCServer) ListTopics(ctx context.Context, req *kafkav1.ListTopicRequest) (*kafkav1.ListTopicReply, error) {
	topics, err := s.Impl.ListTopics(ctx, req.Cluster)
	return &kafkav1.ListTopicReply{Topics: topics}, err
}

// DescribeTopic returns details for a Kafka Topic in the current Kafka Cluster context
func (s *GRPCServer) DescribeTopic(ctx context.Context, req *kafkav1.DescribeTopicRequest) (*kafkav1.DescribeTopicReply, error) {
	topicDescription, err := s.Impl.DescribeTopic(ctx, req.Cluster, req.Topic)
	return &kafkav1.DescribeTopicReply{Topic: topicDescription}, err
}

// CreateTopic creates a new Kafka Topic in the current Kafka Cluster context
func (s *GRPCServer) CreateTopic(ctx context.Context, req *kafkav1.CreateTopicRequest) (*kafkav1.CreateTopicReply, error) {
	return &kafkav1.CreateTopicReply{}, s.Impl.CreateTopic(ctx, req.Cluster, req.Topic)
}

// DeleteTopic deletes a Kafka Topic in the current Kafka Cluster context
func (s *GRPCServer) DeleteTopic(ctx context.Context, req *kafkav1.DeleteTopicRequest) (*kafkav1.DeleteTopicReply, error) {
	return new(kafkav1.DeleteTopicReply), s.Impl.DeleteTopic(ctx, req.Cluster, req.Topic)
}

// UpdateTopic updates any existing Topic's configuration in the current Kafka Cluster context
func (s *GRPCServer) UpdateTopic(ctx context.Context, req *kafkav1.UpdateTopicRequest) (*kafkav1.UpdateTopicReply, error) {
	return new(kafkav1.UpdateTopicReply), s.Impl.UpdateTopic(ctx, req.Cluster, req.Topic)
}

// ListACL lists all ACLs for a given principal or resource
func (s *GRPCServer) ListACL(ctx context.Context, req *kafkav1.ListACLRequest) (*kafkav1.ListACLReply, error) {
	bindings, err := s.Impl.ListACL(ctx, req.Cluster, req.Filter)
	return &kafkav1.ListACLReply{Results: bindings}, err
}

// CreateACL registers a new ACL with the currently Kafka Cluster context
func (s *GRPCServer) CreateACL(ctx context.Context, req *kafkav1.CreateACLRequest) (*kafkav1.CreateACLReply, error) {
	return &kafkav1.CreateACLReply{}, s.Impl.CreateACL(ctx, req.Cluster, req.AclBindings)
}

// DeleteACL removes an ACL with the currently Kafka Cluster context
func (s *GRPCServer) DeleteACL(ctx context.Context, req *kafkav1.DeleteACLRequest) (*kafkav1.DeleteACLReply, error) {
	return &kafkav1.DeleteACLReply{}, s.Impl.DeleteACL(ctx, req.Cluster, req.Filter)
}
