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
		if cfg, err := config.Context(); err == nil {
			client.Kafka.ConfigureKafkaAPI(cfg.Kafka,
				config.Platforms[cfg.Platform].KafkaClusters[cfg.Kafka].APIEndpoint)
		}
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
// ListTopic lists all non-internal topics in the current Kafka cluster context
func (c *Kafka) ListTopic(ctx context.Context) (*kafka.ListKafkaTopicReply, error) {
	c.Logger.Log("msg", "kafka.ListTopic()")

	topics, err := c.Client.Kafka.ListTopic()
	if err != nil {
		return nil, err
	}

	topicList := &kafka.ListKafkaTopicReply{}
	for _, topic := range topics {
		topicList.Topics = append(topicList.Topics, topic.Name)
	}
	return topicList, err
}

// CreateTopic creates a new Kafka Topic in the current Kafka Cluster context
func (c *Kafka) CreateTopic(ctx context.Context, conf *kafka.KafkaAPITopicRequest) (*kafka.KafkaAPIResponse, error) {
	c.Logger.Log("msg", "kafka.CreateTopic()")
	return &kafka.KafkaAPIResponse{}, c.Client.Kafka.CreateTopic(conf)
}

// DescribeTopic returns details for a Kafka Topic in the current Kafka Cluster context
func (c *Kafka) DescribeTopic(ctx context.Context, conf *kafka.KafkaAPITopicRequest) (*kafka.KafkaTopicDescription, error) {
	c.Logger.Log("msg", fmt.Sprintf("kafka.DescribeTopic(%s)", conf.Spec.Name))
	return c.Client.Kafka.DescribeTopic(conf)
}

// DeleteTopic deletes a Kafka Topic in the current Kafka Cluster context
func (c *Kafka) DeleteTopic(ctx context.Context, conf *kafka.KafkaAPITopicRequest) (*kafka.KafkaAPIResponse, error) {
	c.Logger.Log("msg", "kafka.DeleteTopic()")
	return &kafka.KafkaAPIResponse{}, c.Client.Kafka.DeleteTopic(conf)
}

// UpdateTopic updates any existing Topic's configuration in the current Kafka Cluster context
func (c *Kafka) UpdateTopic(ctx context.Context, conf *kafka.KafkaAPITopicRequest) (*kafka.KafkaAPIResponse, error) {
	c.Logger.Log("msg", "kafka.Update")
	return &kafka.KafkaAPIResponse{}, c.Client.Kafka.UpdateTopic(conf)
}

// ListACL registers a new ACL with the currently Kafka cluster context
func (c *Kafka) ListACL(ctx context.Context, conf *kafka.KafkaAPIACLFilterRequest) (*kafka.KafkaAPIACLFilterReply, error) {
	c.Logger.Log("msg", "kafka.ListACL()")
	return c.Client.Kafka.ListACL(conf)
}

// CreateACL registers a new ACL with the currently Kafka Cluster context
func (c *Kafka) CreateACL(ctx context.Context, conf *kafka.KafkaAPIACLRequest) (*kafka.KafkaAPIResponse, error) {
	c.Logger.Log("msg", "kafka.CreateACL()")
	return &kafka.KafkaAPIResponse{}, c.Client.Kafka.CreateACL(conf)
}

// DeleteACL registers a new ACL with the currently Kafka Cluster context
func (c *Kafka) DeleteACL(ctx context.Context, conf *kafka.KafkaAPIACLFilterRequest) (*kafka.KafkaAPIResponse, error) {
	c.Logger.Log("msg", "kafka.DeleteACL()")
	return &kafka.KafkaAPIResponse{}, c.Client.Kafka.DeleteACL(conf)

}

func check(err error) {
	if err != nil {
		golog.Fatal(err)
	}
}
