package main

import (
	"context"
	golog "log"
	"os"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	chttp "github.com/confluentinc/cli/http"
	log "github.com/confluentinc/cli/log"
	metric "github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/ksql"

)

func main() {
	var logger *log.Logger
	{
		logger = log.New()
		logger.Log("msg", "Instantiating plugin " + ksql.Name)
		defer logger.Log("msg", "Shutting down plugin" + ksql.Name)

		f, err := os.OpenFile("/tmp/" + ksql.Name + ".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
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

	var impl *Ksql
	{
		client := chttp.NewClientWithJWT(context.Background(), config.AuthToken, config.AuthURL, config.Logger)
		impl = &Ksql{Logger: logger, Client: client}
	}

	shared.PluginMap[ksql.Name] = &ksql.Plugin{Impl: impl}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: shared.PluginMap,
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

type Ksql struct {
	Logger *log.Logger
	Client *chttp.Client
}

func (c *Ksql) List(ctx context.Context, cluster *schedv1.KSQLCluster) ([]*schedv1.KSQLCluster, error) {
	c.Logger.Log("msg", "ksql.List()")
	ret, _, err := c.Client.Ksql.List(cluster)
	return ret, shared.ConvertAPIError(err)
}

func (c *Ksql) Describe(ctx context.Context, cluster *schedv1.KSQLCluster) (*schedv1.KSQLCluster, error) {
	c.Logger.Log("msg", "ksql.Describe()")
	ret, _, err := c.Client.Ksql.Describe(cluster)
	return ret, shared.ConvertAPIError(err)
}

func (c *Ksql) Create(ctx context.Context, config *schedv1.KSQLClusterConfig) (*schedv1.KSQLCluster, error) {
	c.Logger.Log("msg", "ksql.Create()")
	ret, _, err := c.Client.Ksql.Create(config)
	return ret, shared.ConvertAPIError(err)
}

func (c *Ksql) Delete(ctx context.Context, cluster *schedv1.KSQLCluster) error {
	c.Logger.Log("msg", "ksql.Delete()")
	_, err := c.Client.Ksql.Delete(cluster)
	return shared.ConvertAPIError(err)
}

func check(err error) {
	if err != nil {
		golog.Fatal(err)
	}
}
