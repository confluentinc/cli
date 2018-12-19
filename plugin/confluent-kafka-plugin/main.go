package main

import (
	"context"
	"fmt"
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
		defer logger.Log("msg", "Shutting down plugin"+kafka.Name)

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

func (c *Kafka) CreateAPIKey(ctx context.Context, apiKey *authv1.APIKey) (*authv1.APIKey, error) {
	apiKey, err := c.Client.APIKey.Create(ctx, apiKey)
	return apiKey, shared.ConvertAPIError(err)
}

func (c *Kafka) List(ctx context.Context, cluster *kafkav1.Cluster) ([]*kafkav1.Cluster, error) {
	c.Logger.Log("msg", fmt.Sprintf("kafka.List()\n%+v\n", cluster))
	ret, err := c.Client.Kafka.List(ctx, cluster)
	return ret, shared.ConvertAPIError(err)
}

func (c *Kafka) Describe(ctx context.Context, cluster *kafkav1.Cluster) (*kafkav1.Cluster, error) {
	c.Logger.Log("msg", "kafkav1.Describe()")
	ret, err := c.Client.Kafka.Describe(ctx, cluster)
	return ret, shared.ConvertAPIError(err)
}

func (c *Kafka) Create(ctx context.Context, config *kafkav1.ClusterConfig) (*kafkav1.Cluster, error) {
	c.Logger.Log("msg", "kafkav1.Create()")
	ret, err := c.Client.Kafka.Create(ctx, config)
	return ret, shared.ConvertAPIError(err)
}

func (c *Kafka) Delete(ctx context.Context, cluster *kafkav1.Cluster) error {
	c.Logger.Log("msg", "kafkav1.Delete()")
	return shared.ConvertAPIError(c.Client.Kafka.Delete(ctx, cluster))
}

// ListTopics lists all non-internal topics in the current Kafka cluster context
func (c *Kafka) ListTopics(ctx context.Context, cluster *kafkav1.Cluster) ([]*kafkav1.TopicDescription, error) {
	c.Logger.Log("msg", "kafkav1.ListTopics()")
	return c.Client.Kafka.ListTopics(ctx, cluster)
}

// DescribeTopic returns details for a Kafka Topic in the current Kafka Cluster context
func (c *Kafka) DescribeTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) (*kafkav1.TopicDescription, error) {
	c.Logger.Log("msg", fmt.Sprintf("kafkav1.DescribeTopic(%s)", topic.Spec.Name))
	return c.Client.Kafka.DescribeTopic(ctx, cluster, topic)
}

// CreateTopic creates a new Kafka Topic in the current Kafka Cluster context
func (c *Kafka) CreateTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) error {
	c.Logger.Log("msg", fmt.Sprintf("kafkav1.CreateTopic(%s)", topic.Spec.Name))
	return c.Client.Kafka.CreateTopic(ctx, cluster, topic)
}

// DeleteTopic deletes a Kafka Topic in the current Kafka Cluster context
func (c *Kafka) DeleteTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) error {
	c.Logger.Log("msg", fmt.Sprintf("kafkav1.DeleteTopic(%s)", topic.Spec.Name))
	return c.Client.Kafka.DeleteTopic(ctx, cluster, topic)
}

// ListTopic lists Kafka Topic topic's configuration. This is not implemented in the current version of the CLI
func (c *Kafka) ListTopicConfig(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) (*kafkav1.TopicConfig, error) {
	return nil, shared.ErrNotImplemented
}

// UpdateTopic updates any existing Topic's configuration in the current Kafka Cluster context
func (c *Kafka) UpdateTopic(ctx context.Context, cluster *kafkav1.Cluster, topic *kafkav1.Topic) error {
	c.Logger.Log("msg", fmt.Sprintf("kafkav1.UpdateTopic(%s)", topic.Spec.Name))
	return c.Client.Kafka.UpdateTopic(ctx, cluster, topic)
}

// ListACL registers a new ACL with the currently Kafka cluster context
func (c *Kafka) ListACL(ctx context.Context, cluster *kafkav1.Cluster, filter *kafkav1.ACLFilter) ([]*kafkav1.ACLBinding, error) {
	c.Logger.Log("msg", "kafkav1.ListACL()")
	return c.Client.Kafka.ListACL(ctx, cluster, filter)
}

// CreateACL registers a new ACL with the currently Kafka Cluster context
func (c *Kafka) CreateACL(ctx context.Context, cluster *kafkav1.Cluster, binding []*kafkav1.ACLBinding) error {
	c.Logger.Log("msg", "kafkav1.CreateACL()")
	return c.Client.Kafka.CreateACL(ctx, cluster, binding)
}

// DeleteACL registers a new ACL with the currently Kafka Cluster context
func (c *Kafka) DeleteACL(ctx context.Context, cluster *kafkav1.Cluster, filter *kafkav1.ACLFilter) error {
	c.Logger.Log("msg", "kafkav1.DeleteACL()")
	return c.Client.Kafka.DeleteACL(ctx, cluster, filter)
}

func check(err error) {
	if err != nil {
		golog.Fatal(err)
	}
}
