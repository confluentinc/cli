package main

import (
	"context"
	"os"

	plugin "github.com/hashicorp/go-plugin"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	"github.com/confluentinc/cli/command"
	log "github.com/confluentinc/cli/log"
	metric "github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/apikey"
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
var _ chttp.APIKey = (*ApiKey)(nil)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version") {
		shared.PrintVersion(cliVersion.NewVersion(version, commit, date, host), command.NewTerminalPrompt(os.Stdin))
	}

	var logger *log.Logger
	{
		logger = log.NewWithParams(&log.Params{
			// Plugins log everything. The driver decides the logging level to keep.
			Level:  log.TRACE,
			Output: os.Stderr,
			JSON:   true,
		})
		defer logger.Log("msg", "goodbye")
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

	var impl *ApiKey
	{
		client := chttp.NewClientWithJWT(context.Background(), config.AuthToken, config.AuthURL, config.Logger)
		impl = &ApiKey{Logger: logger, Client: client}
	}

	shared.PluginMap[apikey.Name] = &apikey.Plugin{Impl: impl}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}

type ApiKey struct {
	Logger *log.Logger
	Client *chttp.Client
}

func (c *ApiKey) Create(ctx context.Context, key *authv1.ApiKey) (*authv1.ApiKey, error) {
	c.Logger.Log("msg", "apiKey.Create()")
	ret, err := c.Client.APIKey.Create(ctx, key)
	return ret, shared.ConvertAPIError(err)
}

func (c *ApiKey) Delete(ctx context.Context, key *authv1.ApiKey) error {
	c.Logger.Log("msg", "apiKey.Delete()")
	err := c.Client.APIKey.Delete(ctx, key)
	return shared.ConvertAPIError(err)
}

func (c *ApiKey) List(ctx context.Context, key *authv1.ApiKey) ([]*authv1.ApiKey, error) {
	c.Logger.Log("msg", "apiKey.List()")
	ret, err := c.Client.APIKey.List(ctx, key)
	return ret, shared.ConvertAPIError(err)
}
