package main

import (
	"context"
	"os"

	plugin "github.com/hashicorp/go-plugin"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/command"
	log "github.com/confluentinc/cli/log"
	metric "github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/environment"
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
var _ chttp.Account = (*Account)(nil)

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

	var impl *Account
	{
		client := chttp.NewClientWithJWT(context.Background(), config.AuthToken, config.AuthURL, config.Logger)
		impl = &Account{Logger: logger, Client: client}
	}

	shared.PluginMap[environment.Name] = &environment.Plugin{Impl: impl}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}

type Account struct {
	Logger *log.Logger
	Client *chttp.Client
}

func (c *Account) Create(ctx context.Context, account *orgv1.Account) (*orgv1.Account, error) {
	c.Logger.Log("msg", "Environment.Create()")
	ret, err := c.Client.Account.Create(ctx, account)
	return ret, shared.ConvertAPIError(err)
}

func (c *Account) Update(ctx context.Context, account *orgv1.Account) error {
	c.Logger.Log("msg", "Environment.Update()")
	err := c.Client.Account.Update(ctx, account)
	return shared.ConvertAPIError(err)
}

func (c *Account) Delete(ctx context.Context, account *orgv1.Account) error {
	c.Logger.Log("msg", "Environment.Delete()")
	err := c.Client.Account.Delete(ctx, account)
	return shared.ConvertAPIError(err)
}

func (c *Account) Get(ctx context.Context, account *orgv1.Account) (*orgv1.Account, error) {
	c.Logger.Log("msg", "Environment.Get()")
	ret, err := c.Client.Account.Get(ctx, account)
	return ret, shared.ConvertAPIError(err)
}

func (c *Account) List(ctx context.Context, account *orgv1.Account) ([]*orgv1.Account, error) {
	c.Logger.Log("msg", "Environment.List()")
	ret, err := c.Client.Account.List(ctx, account)
	return ret, shared.ConvertAPIError(err)
}
