package main

import (
	"context"
	golog "log"
	"os"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	log "github.com/confluentinc/cli/log"
	metric "github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/kafka"
)

// Compile-time check for Interface adherence
var _ chttp.Kafka = (*Kafka)(nil)

func main() {
	var logger *log.Logger
	{
		logger = log.New()
		logger.Log("msg", "Instantiating plugin "+kafka.Name)
		defer logger.Log("msg", "Shutting down plugin "+kafka.Name)

		f, err := os.OpenFile("/tmp/"+kafka.Name+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		check(err)
		logger.SetLevel(logrus.DebugLevel)
		logger.Logger.Out = f
	}

	var metricSink shared.MetricSink
	{
		metricSink = metric.NewSink()
	}

	var config *shared.Config
	{
		config = shared.NewConfig(&shared.Config{
			MetricSink: metricSink,
			Logger:     logger,
		})
		err := config.Load()
		if err != nil && err != shared.ErrNoConfig {
			logger.WithError(err).Errorf("unable to load config")
		}
	}

	var impl *Kafka
	{
		client := chttp.NewClientWithJWT(context.Background(), config.AuthToken, config.AuthURL, config.Logger)
		impl = &Kafka{Logger: logger, Client: client}
	}

	shared.PluginMap[kafka.Name] = &kafka.Plugin{Impl: impl}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}

type Kafka struct {
	Logger *log.Logger
	Client *chttp.Client
}

// CreateAPIKey generates an api key for a user
func (c *Kafka) CreateAPIKey(ctx context.Context, apiKey *authv1.ApiKey) (*authv1.ApiKey, error) {
	c.Logger.Log("method", "create", "resource", "api-key",
		"user", apiKey.UserId)
	apiKey, err := c.Client.APIKey.Create(ctx, apiKey)
	return apiKey, shared.ConvertAPIError(err)
}

// List lists the clusters associated with an account
func (c *Kafka) List(ctx context.Context, cluster *kafkav1.KafkaCluster) ([]*kafkav1.KafkaCluster, error) {
	c.Logger.Log(withClusterFields("list", cluster)...)
	ret, err := c.Client.Kafka.List(ctx, cluster)
	return ret, shared.ConvertAPIError(err)
}

// Describe returns details about a particular cluster
func (c *Kafka) Describe(ctx context.Context, cluster *kafkav1.KafkaCluster) (*kafkav1.KafkaCluster, error) {
	c.Logger.Log(withClusterFields("describe", cluster)...)

	ret, err := c.Client.Kafka.Describe(ctx, cluster)
	return ret, shared.ConvertAPIError(err)
}

// Create creates a new cluster
func (c *Kafka) Create(ctx context.Context, config *kafkav1.KafkaClusterConfig) (*kafkav1.KafkaCluster, error) {
	c.Logger.Log(withClusterFields("create",
		&kafkav1.KafkaCluster{AccountId: config.AccountId, Name: config.Name})...)

	ret, err := c.Client.Kafka.Create(ctx, config)
	return ret, shared.ConvertAPIError(err)
}

// Delete destroys a particular cluster from the specified account
func (c *Kafka) Delete(ctx context.Context, cluster *kafkav1.KafkaCluster) error {
	c.Logger.Log(withClusterFields("delete", cluster)...)
	return shared.ConvertAPIError(c.Client.Kafka.Delete(ctx, cluster))
}

// ListTopics lists all non-internal topics in the current Kafka cluster context
func (c *Kafka) ListTopics(ctx context.Context, cluster *kafkav1.KafkaCluster) ([]*kafkav1.TopicDescription, error) {
	c.Logger.Log(withTopicFields("list", cluster, nil)...)

	ret, err := c.Client.Kafka.ListTopics(ctx, cluster)
	return ret, shared.ConvertAPIError(err)
}

// DescribeTopic returns details for a Kafka Topic in the current Kafka Cluster context
func (c *Kafka) DescribeTopic(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) (*kafkav1.TopicDescription, error) {
	c.Logger.Log(withTopicFields("describe", cluster, topic)...)

	ret, err := c.Client.Kafka.DescribeTopic(ctx, cluster, topic)
	return ret, shared.ConvertAPIError(err)
}

// CreateTopic creates a new Kafka Topic in the current Kafka Cluster context
func (c *Kafka) CreateTopic(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) error {
	c.Logger.Log(withTopicFields("create", cluster, topic)...)

	return shared.ConvertAPIError(c.Client.Kafka.CreateTopic(ctx, cluster, topic))
}

// DeleteTopic deletes a Kafka Topic in the current Kafka Cluster context
func (c *Kafka) DeleteTopic(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) error {
	c.Logger.Log(withTopicFields("delete", cluster, topic)...)
	return shared.ConvertAPIError(c.Client.Kafka.DeleteTopic(ctx, cluster, topic))
}

// ListTopicConfig lists Kafka Topic topic's configuration. This is not implemented in the current version of the CLI
func (c *Kafka) ListTopicConfig(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) (*kafkav1.TopicConfig, error) {
	return nil, shared.ErrNotImplemented
}

// UpdateTopic updates any existing Topic's configuration in the current Kafka Cluster context
func (c *Kafka) UpdateTopic(ctx context.Context, cluster *kafkav1.KafkaCluster, topic *kafkav1.Topic) error {
	c.Logger.Log(withTopicFields("update", cluster, topic)...)
	return shared.ConvertAPIError(c.Client.Kafka.UpdateTopic(ctx, cluster, topic))
}

// ListACL registers a new ACL with the currently Kafka cluster context
func (c *Kafka) ListACL(ctx context.Context, cluster *kafkav1.KafkaCluster, filter *kafkav1.ACLFilter) ([]*kafkav1.ACLBinding, error) {
	c.Logger.Log(withACLFields("list", cluster, filter.PatternFilter)...)
	ret, err := c.Client.Kafka.ListACL(ctx, cluster, filter)
	return ret, shared.ConvertAPIError(err)
}

// CreateACL registers a new ACL with the currently Kafka Cluster context
func (c *Kafka) CreateACL(ctx context.Context, cluster *kafkav1.KafkaCluster, binding []*kafkav1.ACLBinding) error {
	c.Logger.Log(withACLFields("create", cluster, binding[0].Pattern)...)

	return shared.ConvertAPIError(c.Client.Kafka.CreateACL(ctx, cluster, binding))
}

// DeleteACL registers a new ACL with the currently Kafka Cluster context
func (c *Kafka) DeleteACL(ctx context.Context, cluster *kafkav1.KafkaCluster, filter *kafkav1.ACLFilter) error {
	c.Logger.Log(withACLFields("delete", cluster, filter.PatternFilter)...)
	return shared.ConvertAPIError(c.Client.Kafka.DeleteACL(ctx, cluster, filter))
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
	fields := []interface{}{"method", method, "resource", resource}

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

func check(err error) {
	if err != nil {
		golog.Fatal(err)
	}
}
