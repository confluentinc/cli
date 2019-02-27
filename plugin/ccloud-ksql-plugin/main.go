package main

import (
	"context"
	"os"

	plugin "github.com/hashicorp/go-plugin"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	ksqlv1 "github.com/confluentinc/ccloudapis/ksql/v1"
	"github.com/confluentinc/cli/command"
	log "github.com/confluentinc/cli/log"
	metric "github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/ksql"
	cliVersion "github.com/confluentinc/cli/version"
)

var (
	// Injected from linker flags like `go build -ldflags "-X main.version=$VERSION" -X ...`
	version = "v0.0.0"
	commit  = ""
	date    = ""
	host    = ""
)

func main() {
	if os.Args[1] == "version" || os.Args[1] == "--version" {
		shared.PrintVersion(cliVersion.NewVersion(version, commit, date, host), command.NewTerminalPrompt(os.Stdin))
	}

	var logger *log.Logger
	{
		logger = log.New()
		logger.Log("msg", "hello")
		defer logger.Log("msg", "goodbye")

		f, err := os.OpenFile("/tmp/confluent-ksql-plugin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		check(err, logger)
		logger.SetLevel(log.DEBUG)
		logger.SetOutput(f)
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
			logger.Errorf("unable to load config: %v", err)
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
		Plugins:         shared.PluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}

type Ksql struct {
	Logger *log.Logger
	Client *chttp.Client
}

func (c *Ksql) List(ctx context.Context, cluster *ksqlv1.KSQLCluster) ([]*ksqlv1.KSQLCluster, error) {
	c.Logger.Log("msg", "ksql.List()")
	ret, err := c.Client.KSQL.List(ctx, cluster)
	return ret, shared.ConvertAPIError(err)
}

func (c *Ksql) Describe(ctx context.Context, cluster *ksqlv1.KSQLCluster) (*ksqlv1.KSQLCluster, error) {
	c.Logger.Log("msg", "ksql.Describe()")
	ret, err := c.Client.KSQL.Describe(ctx, cluster)
	return ret, shared.ConvertAPIError(err)
}

func (c *Ksql) Create(ctx context.Context, config *ksqlv1.KSQLClusterConfig) (*ksqlv1.KSQLCluster, error) {
	c.Logger.Log("msg", "ksql.Create()")
	ret, err := c.Client.KSQL.Create(ctx, config)
	return ret, shared.ConvertAPIError(err)
}

func (c *Ksql) Delete(ctx context.Context, cluster *ksqlv1.KSQLCluster) error {
	c.Logger.Log("msg", "ksql.Delete()")
	err := c.Client.KSQL.Delete(ctx, cluster)
	return shared.ConvertAPIError(err)
}

func check(err error, logger *log.Logger) {
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
