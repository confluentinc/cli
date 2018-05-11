package main

import (
	"context"
	golog "log"
	"os"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/command/kafka"
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

		f, err := os.OpenFile("/tmp/confluent-kafka-plugin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
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

	var impl *Kafka
	{
		client := chttp.NewClientWithJWT(context.Background(), config.AuthToken, config.AuthURL, config.Logger)
		impl = &Kafka{Logger: logger, Client: client}
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"kafka": &kafka.Plugin{Impl: impl},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

type Kafka struct {
	Logger *log.Logger
	Client *chttp.Client
}

func (c *Kafka) List(ctx context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error) {
	c.Logger.Log("msg", "kafka.List()")
	ret, _, err := c.Client.Kafka.List(cluster)
	c.Logger.Log("msg", ret)
	return ret, shared.ConvertAPIError(err)
}

func (c *Kafka) Describe(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
	c.Logger.Log("msg", "kafka.Describe()")
	ret, _, err := c.Client.Kafka.Describe(cluster)
	return ret, shared.ConvertAPIError(err)
}

func check(err error) {
	if err != nil {
		golog.Fatal(err)
	}
}
