package main

import (
	"context"
	golog "log"
	"os"

	"github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/command/connect"
	chttp "github.com/confluentinc/cli/http"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
	proto "github.com/confluentinc/cli/shared/connect"
)

func main() {
	var logger *log.Logger
	{
		logger = log.New()
		logger.Log("msg", "hello")
		defer logger.Log("msg", "goodbye")

		f, err := os.OpenFile("/tmp/confluent-connect-plugin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
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

func (c *Connect) CreateS3Sink(ctx context.Context, cfg *proto.ConnectS3SinkClusterConfig) (*schedv1.ConnectS3SinkCluster, error) {
	c.Logger.Log("msg", "connect.CreateS3Sink()")
	config := &schedv1.ConnectS3SinkClusterConfig{
		Name:           cfg.Name,
		AccountId:      cfg.AccountId,
		KafkaClusterId: cfg.KafkaClusterId,
		Servers:        cfg.Servers,
		Options:        cfg.Options,
	}

	// Resolve kafka user email -> ID
	user, _, err := c.Client.User.Describe(&orgv1.User{Email: cfg.KafkaUserEmail})
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}
	config.KafkaUserId = user.Id

	// Create the connect cluster
	ret, _, err := c.Client.Connect.CreateS3Sink(config)
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}

	return ret, nil
}

func (c *Connect) UpdateS3Sink(ctx context.Context, cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, error) {
	c.Logger.Log("msg", "connect.UpdateS3Sink()")
	cluster, _, err := c.Client.Connect.UpdateS3Sink(cluster)
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}

	return cluster, nil
}

func (c *Connect) Delete(ctx context.Context, cluster *schedv1.ConnectCluster) error {
	c.Logger.Log("msg", "connect.Delete()")
	_, err := c.Client.Connect.Delete(cluster)
	if err != nil {
		return shared.ConvertAPIError(err)
	}
	return nil
}

func check(err error) {
	if err != nil {
		golog.Fatal(err)
	}
}
