package main

import (
	"context"
	golog "log"
	"os"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"

	chttp "github.com/confluentinc/cli/http"
	log "github.com/confluentinc/cli/log"
	metric "github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/command/api-key"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
)

func main() {
	var logger *log.Logger
	{
		logger = log.New()
		logger.Log("msg", "hello")
		defer logger.Log("msg", "goodbye")

		f, err := os.OpenFile("/tmp/confluent-api-key-plugin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
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

	var impl *ApiKey
	{
		client := chttp.NewClientWithJWT(context.Background(), config.AuthToken, config.AuthURL, config.Logger)
		impl = &ApiKey{Logger: logger, Client: client}
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"apiKey": &apiKey.Plugin{Impl: impl},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

type ApiKey struct {
	Logger *log.Logger
	Client *chttp.Client
}

func (c *ApiKey) Create(ctx context.Context, key *schedv1.ApiKey) (*schedv1.ApiKey, error) {
	c.Logger.Log("msg", "apiKey.Create()")
	ret, _, err := c.Client.APIKey.Create(key)
	return ret, shared.ConvertAPIError(err)
}

func (c *ApiKey) Delete(ctx context.Context,  key *schedv1.ApiKey) error {
	c.Logger.Log("msg", "apiKey.Delete()")
	_, err :=  c.Client.APIKey.Delete(key)
	return shared.ConvertAPIError(err)
}

func check(err error) {
	if err != nil {
		golog.Fatal(err)
	}
}

