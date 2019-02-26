package main

import (
	"context"
	golog "log"
	"os"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/user"
)

// Compile-time check for Interface adherence
var _ chttp.User = (*User)(nil)

func main() {
	var logger *log.Logger
	{
		logger = log.New()
		logger.Log("msg", "hello")
		defer logger.Log("msg", "goodbye")

		f, err := os.OpenFile("/tmp/confluent-user-plugin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
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

func check(err error) {
	if err != nil {
		golog.Fatal(err)
	}
}
