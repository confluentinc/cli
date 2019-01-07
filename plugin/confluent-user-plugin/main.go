package main

import (
	"context"
	golog "log"
	"os"
	"plugin"

	"github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/cli/command/user"
	chttp "github.com/confluentinc/cli/ccloud-sdk-go"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/metric"
	"github.com/confluentinc/cli/shared"
)

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

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"user": &user.Plugin{Impl: impl},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

type User struct {
	Logger *log.Logger
	Client *chttp.Client
}

func (c *User) CreateServiceAccount(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
	c.Logger.Log("msg", "user.UpdateServiceAccount()")
	ret, _, err := c.Client.User.CreateServiceAccount(user)
	/*if err != nil && err != shared.ErrNoConfig {
		c.Logger.Log("err", "Errorr")
	}*/
	c.Logger.Log("msg", "return val:::")
	return ret, shared.ConvertAPIError(err)
}

func (c *User) UpdateServiceAccount(ctx context.Context, user *orgv1.User) error {
	c.Logger.Log("msg", "user.UpdateServiceAccount()")
	_, err := c.Client.User.UpdateServiceAccount(user)
	c.Logger.Log("msg", "return val:::")
	return shared.ConvertAPIError(err)
}

func (c *User) DeactivateServiceAccount(ctx context.Context, user *orgv1.User) error {
	c.Logger.Log("msg", "user.DeactivateServiceAccount()")
	_, err := c.Client.User.DeactivateServiceAccount(user)
	return shared.ConvertAPIError(err)
}

func (c *User) GetServiceAccounts(ctx context.Context, user *orgv1.User) ([]*orgv1.User, error) {
	c.Logger.Log("msg", "user.CreateServiceAccount()")
	ret, _, err := c.Client.User.GetServiceAccounts(user)
	return ret, shared.ConvertAPIError(err)
}

func check(err error) {
	if err != nil {
		golog.Fatal(err)
	}
}
