package main

import (
	"context"
	golog "log"
	"os"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/command/connect"
	chttp "github.com/confluentinc/cli/http"
	log "github.com/confluentinc/cli/log"
	metric "github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
)

func main() {
	var logger *log.Logger
	{
		logger = log.New()
		logger.Log("msg", "hello")
		defer logger.Log("msg", "goodbye")

		f, err := os.OpenFile("/tmp/confluent-connect-plugin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
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
		config = &shared.Config{
			MetricSink: metricSink,
			Logger:     logger,
		}
		err := config.Load()
		if err != nil && err != shared.ErrNoConfig {
			logger.WithError(err).Errorf("unable to load config")
		}
	}

	var impl *Connect
	{
		client := chttp.NewClientWithJWT(context.Background(), config.AuthToken, config.AuthURL, config.Logger)
		impl = &Connect{Logger: logger, Client: client}
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"connect": &connect.Plugin{Impl: impl},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

type Connect struct {
	Logger *log.Logger
	Client *chttp.Client
}

func (c *Connect) List(ctx context.Context, cluster *schedv1.ConnectCluster) ([]*schedv1.ConnectCluster, error) {
	c.Logger.Log("msg", "connect.List()")
	ret, _, err := c.Client.Connect.List(cluster)
	return ret, shared.ConvertAPIError(err)
}

func (c *Connect) Describe(ctx context.Context, cluster *schedv1.ConnectCluster) (*schedv1.ConnectCluster, error) {
	c.Logger.Log("msg", "connect.Describe()")
	ret, _, err := c.Client.Connect.Describe(cluster)
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}
	return ret, nil
}

func (c *Connect) DescribeS3Sink(ctx context.Context, cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, error) {
	c.Logger.Log("msg", "connect.DescribeS3Sink()")
	ret, _, err := c.Client.Connect.DescribeS3Sink(cluster)
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}
	return ret, nil
}

func (c *Connect) CreateS3Sink(ctx context.Context, cfg *schedv1.ConnectS3SinkClusterConfig) (*schedv1.ConnectS3SinkCluster, error) {
	c.Logger.Log("msg", "connect.CreateS3Sink()")

	// Resolve kafka name -> ID
	kafkaName := cfg.KafkaClusterId // NOTE: the CLI stored the name in KafkaClusterId field
	kafka, _, err := c.Client.Kafka.Describe(&schedv1.KafkaCluster{AccountId: cfg.AccountId, Name: kafkaName})
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}
	cfg.KafkaClusterId = kafka.Id

	// Create the connect cluster
	ret, _, err := c.Client.Connect.CreateS3Sink(cfg)
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}
	ret.KafkaClusterId = kafkaName  // NOTE: store the Name in the ID field for the Detail view

	return ret, nil
}

func (c *Connect) UpdateS3Sink(ctx context.Context, cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, error) {
	c.Logger.Log("msg", "connect.UpdateS3Sink()")
	resolved, err := c.resolveConnectClusterID(ctx, cluster.ConnectCluster)
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}
	cluster.Id = resolved.Id

	// Update the connect cluster
	cluster, _, err = c.Client.Connect.UpdateS3Sink(cluster)
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}

	// Store the kafka cluster name back into the ID field for the Detail view
	cluster.KafkaClusterId, err = c.resolveKafkaClusterName(cluster.ConnectCluster)
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}

	return cluster, nil
}

func (c *Connect) Delete(ctx context.Context, cluster *schedv1.ConnectCluster) error {
	c.Logger.Log("msg", "connect.Delete()")
	cluster, err := c.resolveConnectClusterID(ctx, cluster)
	if err != nil {
		return shared.ConvertAPIError(err)
	}
	_, err = c.Client.Connect.Delete(cluster)
	if err != nil {
		return shared.ConvertAPIError(err)
	}
	return nil
}

// resolveConnectClusterID resolves connect name to id
func (c *Connect) resolveConnectClusterID(ctx context.Context, cluster *schedv1.ConnectCluster) (*schedv1.ConnectCluster, error) {
	if cluster.Id != "" {
		return cluster, nil
	}
	cluster, err := c.Describe(ctx, cluster)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

// resolveKafkaClusterName resolves kafka id to name
func (c *Connect) resolveKafkaClusterName(cluster *schedv1.ConnectCluster) (string, error) {
	kafka, _, err := c.Client.Kafka.Describe(&schedv1.KafkaCluster{AccountId: cluster.AccountId, Id: cluster.KafkaClusterId})
	if err != nil {
		return "", shared.ConvertAPIError(err)
	}
	return kafka.Name, nil
}

func check(err error) {
	if err != nil {
		golog.Fatal(err)
	}
}
