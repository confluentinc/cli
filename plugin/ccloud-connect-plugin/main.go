package main

import (
	"context"
	"os"

	"github.com/hashicorp/go-plugin"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	connectv1 "github.com/confluentinc/ccloudapis/connect/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/connect"
	cliVersion "github.com/confluentinc/cli/version"
)

var (
	// Injected from linker flags like `go build -ldflags "-X main.version=$VERSION" -X ...`
	version = "v0.0.0"
	commit  = ""
	date    = ""
	host    = ""
)

// Compile-time check for Interface adherence
var _ chttp.Connect = (*Connect)(nil)

func main() {
	if os.Args[1] == "version" || os.Args[1] == "--version" {
		shared.PrintVersion(cliVersion.NewVersion(version, commit, date, host), command.NewTerminalPrompt(os.Stdin))
	}

	var logger *log.Logger
	{
		logger = log.New()
		logger.Log("msg", "hello")
		defer logger.Log("msg", "goodbye")

		f, err := os.OpenFile("/tmp/confluent-connect-plugin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
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
		config = shared.NewConfig(&shared.Config{
			MetricSink: metricSink,
			Logger:     logger,
		})
		err := config.Load()
		if err != nil && err != shared.ErrNoConfig {
			logger.Errorf("unable to load config: %v", err)
		}
	}

	var impl *Connect
	{
		client := chttp.NewClientWithJWT(context.Background(), config.AuthToken, config.AuthURL, config.Logger)
		impl = &Connect{Logger: logger, Client: client}
	}

	shared.PluginMap[connect.Name] = &connect.Plugin{Impl: impl}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}

type Connect struct {
	Logger *log.Logger
	Client *chttp.Client
}

func (c *Connect) List(ctx context.Context, cluster *connectv1.ConnectCluster) ([]*connectv1.ConnectCluster, error) {
	c.Logger.Log("msg", "connect.List()")
	ret, err := c.Client.Connect.List(ctx, cluster)
	return ret, shared.ConvertAPIError(err)
}

func (c *Connect) Describe(ctx context.Context, cluster *connectv1.ConnectCluster) (*connectv1.ConnectCluster, error) {
	c.Logger.Log("msg", "connect.Describe()")
	ret, err := c.Client.Connect.Describe(ctx, cluster)
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}
	return ret, nil
}

func (c *Connect) DescribeS3Sink(ctx context.Context, cluster *connectv1.ConnectS3SinkCluster) (*connectv1.ConnectS3SinkCluster, error) {
	c.Logger.Log("msg", "connect.DescribeS3Sink()")
	ret, err := c.Client.Connect.DescribeS3Sink(ctx, cluster)
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}
	return ret, nil
}

func (c *Connect) CreateS3Sink(ctx context.Context, cfg *connectv1.ConnectS3SinkClusterConfig) (*connectv1.ConnectS3SinkCluster, error) {
	c.Logger.Log("msg", "connect.CreateS3Sink()")
	config := &connectv1.ConnectS3SinkClusterConfig{
		Name:           cfg.Name,
		AccountId:      cfg.AccountId,
		KafkaClusterId: cfg.KafkaClusterId,
		Servers:        cfg.Servers,
		Options:        cfg.Options,
	}

	// Resolve kafka user email -> ID
	user, err := c.Client.User.Describe(ctx, &orgv1.User{Email: cfg.UserEmail})
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}
	config.KafkaUserId = user.Id

	// Create the connect cluster
	ret, err := c.Client.Connect.CreateS3Sink(ctx, config)
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}

	return ret, nil
}

func (c *Connect) UpdateS3Sink(ctx context.Context, cluster *connectv1.ConnectS3SinkCluster) (*connectv1.ConnectS3SinkCluster, error) {
	c.Logger.Log("msg", "connect.UpdateS3Sink()")
	cluster, err := c.Client.Connect.UpdateS3Sink(ctx, cluster)
	if err != nil {
		return nil, shared.ConvertAPIError(err)
	}

	return cluster, nil
}

func (c *Connect) Delete(ctx context.Context, cluster *connectv1.ConnectCluster) error {
	c.Logger.Log("msg", "connect.Delete()")
	err := c.Client.Connect.Delete(ctx, cluster)
	if err != nil {
		return shared.ConvertAPIError(err)
	}
	return nil
}

func check(err error, logger *log.Logger) {
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
