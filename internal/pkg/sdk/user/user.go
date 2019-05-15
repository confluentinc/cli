package user

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/pkg/log"
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
	return c.Client.User.List(ctx)
}

func (c *User) Describe(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
	c.Logger.Log("msg", "user.Describe()")
	return c.Client.User.Describe(ctx, user)
}

func (c *User) CreateServiceAccount(ctx context.Context, user *orgv1.User) (*orgv1.User, error) {
	c.Logger.Log("msg", "user.CreateServiceAccount()")
	return c.Client.User.CreateServiceAccount(ctx, user)
}

func (c *User) UpdateServiceAccount(ctx context.Context, user *orgv1.User) error {
	c.Logger.Log("msg", "user.UpdateServiceAccount()")
	return c.Client.User.UpdateServiceAccount(ctx, user)
}

func (c *User) DeleteServiceAccount(ctx context.Context, user *orgv1.User) error {
	c.Logger.Log("msg", "user.DeleteServiceAccount()")
	return c.Client.User.DeleteServiceAccount(ctx, user)
}

func (c *User) GetServiceAccounts(ctx context.Context) ([]*orgv1.User, error) {
	c.Logger.Log("msg", "user.GetServiceAccounts()")
	return c.Client.User.GetServiceAccounts(ctx)
}
