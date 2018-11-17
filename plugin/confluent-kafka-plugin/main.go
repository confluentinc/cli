package main

import (
	"context"
	"fmt"
	golog "log"
	"os"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	chttp "github.com/confluentinc/cli/http"
	log "github.com/confluentinc/cli/log"
	metric "github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/kafka"
)

func main() {
	var logger *log.Logger
	{
		logger = log.New()
		logger.Log("msg", "Instantiating plugin " + kafka.Name)
		defer logger.Log("msg", "Shutting down plugin" + kafka.Name)

		f, err := os.OpenFile("/tmp/" + kafka.Name + ".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
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
		Plugins: shared.PluginMap,
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

type Kafka struct {
	Logger *log.Logger
	Client *chttp.Client
}

func (c *Kafka) CreateAPIKey(ctx context.Context, apiKey *schedv1.ApiKey) (*schedv1.ApiKey, error) {
	ret, _, err := c.Client.APIKey.Create(apiKey)
	return ret, shared.ConvertAPIError(err)
}

func (c *Kafka) List(ctx context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error) {
	c.Logger.Log("msg", "kafka.List()")
	ret, _, err := c.Client.Kafka.List(cluster)
	return ret, shared.ConvertAPIError(err)
}

func (c *Kafka) Describe(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
	c.Logger.Log("msg", "kafka.Describe()")
	ret, _, err := c.Client.Kafka.Describe(cluster)
	return ret, shared.ConvertAPIError(err)
}

func (c *Kafka) Create(ctx context.Context, config *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, error) {
	c.Logger.Log("msg", "kafka.Create()")
	ret, _, err := c.Client.Kafka.Create(config)
	return ret, shared.ConvertAPIError(err)
}

func (c *Kafka) Delete(ctx context.Context, cluster *schedv1.KafkaCluster) error {
	c.Logger.Log("msg", "kafka.Delete()")
	_, err := c.Client.Kafka.Delete(cluster)
	return shared.ConvertAPIError(err)
}

// TODO: Move filtering to the "driver" under commnad/kafka return  []KafkaTopicDescription instead

// ListTopics lists all non-internal topics in the current Kafka cluster context
func (c *Kafka) ListTopics(ctx context.Context, cluster *schedv1.KafkaCluster) ([]string, error) {
	c.Logger.Log("msg", "kafka.ListTopics()")
	topics, err := c.Client.Kafka.ListTopics(cluster)

	if err != nil {
		return nil, err
	}

	var topicList []string
	for _, topic := range topics {
		topicList = append(topicList, topic.Name)
	}
	return topicList, err
}

// CreateTopic creates a new Kafka Topic in the current Kafka Cluster context
func (c *Kafka) CreateTopic(ctx context.Context, cluster *schedv1.KafkaCluster, topic *kafka.Topic) (*kafka.KafkaAPIResponse, error) {
	c.Logger.Log("msg", fmt.Sprintf("kafka.CreateTopic(%s)", topic.Spec.Name))
	return &kafka.KafkaAPIResponse{}, c.Client.Kafka.CreateTopic(cluster, topic)
}

// DescribeTopic returns details for a Kafka Topic in the current Kafka Cluster context
func (c *Kafka) DescribeTopic(ctx context.Context, cluster *schedv1.KafkaCluster, topic *kafka.Topic) (*kafka.KafkaTopicDescription, error) {
	c.Logger.Log("msg", fmt.Sprintf("kafka.DescribeTopic(%s)", topic.Spec.Name))
	return c.Client.Kafka.DescribeTopic(cluster, topic)
}

// DeleteTopic deletes a Kafka Topic in the current Kafka Cluster context
func (c *Kafka) DeleteTopic(ctx context.Context, cluster *schedv1.KafkaCluster, topic *kafka.Topic) (*kafka.KafkaAPIResponse, error) {
	c.Logger.Log("msg", fmt.Sprintf("kafka.DeleteTopic(%s)", topic.Spec.Name))
	return &kafka.KafkaAPIResponse{}, c.Client.Kafka.DeleteTopic(cluster, topic)
}

// UpdateTopic updates any existing Topic's configuration in the current Kafka Cluster context
func (c *Kafka) UpdateTopic(ctx context.Context, cluster *schedv1.KafkaCluster, topic *kafka.Topic) (*kafka.KafkaAPIResponse, error) {
	c.Logger.Log("msg", fmt.Sprintf("kafka.DeleteTopic(%s)", topic.Spec.Name))
	return &kafka.KafkaAPIResponse{}, c.Client.Kafka.UpdateTopic(cluster, topic)
}

// ListACL registers a new ACL with the currently Kafka cluster context
func (c *Kafka) ListACL(ctx context.Context, conf *kafka.ACLFilter) (*kafka.KafkaAPIACLFilterReply, error) {
	c.Logger.Log("msg", "kafka.ListACL()")
	return c.Client.Kafka.ListACL(conf)
}

// CreateACL registers a new ACL with the currently Kafka Cluster context
func (c *Kafka) CreateACL(ctx context.Context, conf *kafka.ACLSpec) (*kafka.KafkaAPIResponse, error) {
	c.Logger.Log("msg", "kafka.CreateACL()")
	return &kafka.KafkaAPIResponse{}, c.Client.Kafka.CreateACL(conf)
}

// DeleteACL registers a new ACL with the currently Kafka Cluster context
func (c *Kafka) DeleteACL(ctx context.Context, conf *kafka.ACLFilter) (*kafka.KafkaAPIResponse, error) {
	c.Logger.Log("msg", "kafka.DeleteACL()")
	return &kafka.KafkaAPIResponse{}, c.Client.Kafka.DeleteACL(conf)

}

func check(err error) {
	if err != nil {
		golog.Fatal(err)
	}
}
