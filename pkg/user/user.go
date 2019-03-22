package user

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/log"
	"github.com/confluentinc/cli/shared"
)

// Compile-time check for Interface adherence
var _ ccloud.User = (*User)(nil)

type User struct {
	Client *ccloud.Client
	Logger *log.Logger
}

func New(client *ccloud.Client, logger *log.Logger) *User {
	return &User{Client: client, Logger: logger}
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
