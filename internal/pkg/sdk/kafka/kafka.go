package kafka

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

// Compile-time check for Interface adherence
var _ ccloud.Kafka = (*Kafka)(nil)

type Kafka struct {
	Client *ccloud.Client
	Logger *log.Logger
}

func New(client *ccloud.Client, logger *log.Logger) *Kafka {
	return &Kafka{Client: client, Logger: logger}
}

// CreateAPIKey generates an api key for a user
func (c *Kafka) CreateAPIKey(ctx context.Context, apiKey *authv1.ApiKey) (*authv1.ApiKey, error) {
	c.Logger.Log("method", "create", "resource", "api-key", "user", apiKey.UserId)
	return c.Client.APIKey.Create(ctx, apiKey)
}

// List lists the clusters associated with an account
func (c *Kafka) List(ctx context.Context, cluster *kafkav1.KafkaCluster) ([]*kafkav1.KafkaCluster, error) {
	c.Logger.Log(withClusterFields("list", cluster)...)
	return c.Client.Kafka.List(ctx, cluster)
}

// Describe returns details about a particular cluster
func (c *Kafka) Describe(ctx context.Context, cluster *kafkav1.KafkaCluster) (*kafkav1.KafkaCluster, error) {
	c.Logger.Log(withClusterFields("describe", cluster)...)

	return c.Client.Kafka.Describe(ctx, cluster)
}

// Create creates a new cluster
func (c *Kafka) Create(ctx context.Context, config *kafkav1.KafkaClusterConfig) (*kafkav1.KafkaCluster, error) {
	c.Logger.Log(withClusterFields("create",
		&kafkav1.KafkaCluster{AccountId: config.AccountId, Name: config.Name})...)

	return c.Client.Kafka.Create(ctx, config)
}

// Delete destroys a particular cluster from the specified account
func (c *Kafka) Delete(ctx context.Context, cluster *kafkav1.KafkaCluster) error {
	c.Logger.Log(withClusterFields("delete", cluster)...)
	return c.Client.Kafka.Delete(ctx, cluster)
}

// ListTopics lists all non-internal topics in the current Kafka cluster context
func (c *Kafka) ListTopics(ctx context.Context, cluster *kafkav1.KafkaCluster) ([]*kafkav1.TopicDescription, error) {
	c.Logger.Log(withTopicFields("list", cluster, nil)...)
	return c.Client.Kafka.ListTopics(ctx, cluster)
}

// DescribeTopic returns details for a Kafka Topic in the current Kafka Cluster context
func (c *Kafka) DescribeTopic(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) (*kafkav1.TopicDescription, error) {
	c.Logger.Log(withTopicFields("describe", cluster, topic)...)
	return c.Client.Kafka.DescribeTopic(ctx, cluster, topic)
}

// CreateTopic creates a new Kafka Topic in the current Kafka Cluster context
func (c *Kafka) CreateTopic(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) error {
	c.Logger.Log(withTopicFields("create", cluster, topic)...)
	return c.Client.Kafka.CreateTopic(ctx, cluster, topic)
}

// DeleteTopic deletes a Kafka Topic in the current Kafka Cluster context
func (c *Kafka) DeleteTopic(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) error {
	c.Logger.Log(withTopicFields("delete", cluster, topic)...)
	return c.Client.Kafka.DeleteTopic(ctx, cluster, topic)
}

// ListTopicConfig lists Kafka Topic topic's configuration. This is not implemented in the current version of the CLI
func (c *Kafka) ListTopicConfig(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) (*kafkav1.TopicConfig, error) {
	return nil, errors.ErrNotImplemented
}

// UpdateTopic updates any existing Topic's configuration in the current Kafka Cluster context
func (c *Kafka) UpdateTopic(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) error {
	c.Logger.Log(withTopicFields("update", cluster, topic)...)
	return c.Client.Kafka.UpdateTopic(ctx, cluster, topic)
}

// ListACL registers a new ACL with the currently Kafka cluster context
func (c *Kafka) ListACL(ctx context.Context, cluster *kafkav1.KafkaCluster, filter *kafkav1.ACLFilter) ([]*kafkav1.ACLBinding, error) {
	c.Logger.Log(withACLFields("list", cluster, filter.PatternFilter)...)
	return c.Client.Kafka.ListACL(ctx, cluster, filter)
}

// CreateACL registers a new ACL with the currently Kafka Cluster context
func (c *Kafka) CreateACL(ctx context.Context, cluster *kafkav1.KafkaCluster, binding []*kafkav1.ACLBinding) error {
	c.Logger.Log(withACLFields("create", cluster, binding[0].Pattern)...)
	return c.Client.Kafka.CreateACL(ctx, cluster, binding)
}

// DeleteACL registers a new ACL with the currently Kafka Cluster context
func (c *Kafka) DeleteACL(ctx context.Context, cluster *kafkav1.KafkaCluster, filter *kafkav1.ACLFilter) error {
	c.Logger.Log(withACLFields("delete", cluster, filter.PatternFilter)...)
	return c.Client.Kafka.DeleteACL(ctx, cluster, filter)
}

func withClusterFields(method string, cluster *kafkav1.KafkaCluster) []interface{} {
	return withFields(method, "cluster", cluster, nil, nil)
}

func withTopicFields(method string, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) []interface{} {
	return withFields(method, "topic", cluster, topic, nil)
}

func withACLFields(method string, cluster *kafkav1.KafkaCluster, acl *kafkav1.ResourcePatternConfig) []interface{} {
	return withFields(method, "acl", cluster, nil, acl)
}

func withFields(method string, resource string, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic, acl *kafkav1.ResourcePatternConfig) []interface{} {
	fields := []interface{}{"msg", "request", "method", method, "resource", resource}

	if cluster != nil {
		fields = append(fields, "cluster", cluster.Id, "account", cluster.AccountId)
	}
	if topic != nil {
		fields = append(fields, "name", topic.Spec.Name)
	}
	if acl != nil {
		fields = append(fields, "name", acl.Name, "acl_resource", acl.ResourceType)
	}
	return fields
}
