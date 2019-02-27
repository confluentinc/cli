package main

import (
	"context"
	"os"

	plugin "github.com/hashicorp/go-plugin"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/user"
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
var _ chttp.User = (*User)(nil)

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

	var impl *User
	{
		client := chttp.NewClientWithJWT(context.Background(), config.AuthToken, config.AuthURL, config.Logger)
		impl = &User{Logger: logger, Client: client}
	}

	shared.PluginMap[user.Name] = &user.Plugin{Impl: impl}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}

type User struct {
	Logger *log.Logger
	Client *chttp.Client
}

func (c *User) List(ctx context.Context) ([]*orgv1.User, error) {
	c.Logger.Log("msg", "user.List()")
	ret, err := c.Client.User.List(ctx)
	return ret, shared.ConvertAPIError(err)
}

func (c *User) Describe(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
	c.Logger.Log("msg", "user.Describe()")
	ret, err := c.Client.User.Describe(ctx, user)
	return ret, shared.ConvertAPIError(err)
}

func (c *User) CreateServiceAccount(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
	c.Logger.Log("msg", "user.CreateServiceAccount()")
	ret, err := c.Client.User.CreateServiceAccount(ctx, user)
	return ret, shared.ConvertAPIError(err)
}

func (c *User) UpdateServiceAccount(ctx context.Context, user *orgv1.User) error {
	c.Logger.Log("msg", "user.UpdateServiceAccount()")
	err := c.Client.User.UpdateServiceAccount(ctx, user)
	return shared.ConvertAPIError(err)
}

func (c *User) DeleteServiceAccount(ctx context.Context, user *orgv1.User) error {
	c.Logger.Log("msg", "user.DeleteServiceAccount()")
	err := c.Client.User.DeleteServiceAccount(ctx, user)
	return shared.ConvertAPIError(err)
}

func (c *User) GetServiceAccounts(ctx context.Context) ([]*orgv1.User, error) {
	c.Logger.Log("msg", "user.GetServiceAccounts()")
	ret, err := c.Client.User.GetServiceAccounts(ctx)
	return ret, shared.ConvertAPIError(err)
}
