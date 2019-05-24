package environment

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/pkg/log"
)

// Compile-time check for Interface adherence
var _ ccloud.Account = (*Environment)(nil)

type Environment struct {
	Logger *log.Logger
	Client *ccloud.Client
}

func New(client *ccloud.Client, logger *log.Logger) *Environment {
	return &Environment{Client: client, Logger: logger}
}

func (c *Environment) Create(ctx context.Context, account *orgv1.Account) (*orgv1.Account, error) {
	c.Logger.Log("msg", "Environment.Create()")
	return c.Client.Account.Create(ctx, account)
}

func (c *Environment) Update(ctx context.Context, account *orgv1.Account) error {
	c.Logger.Log("msg", "Environment.Update()")
	return c.Client.Account.Update(ctx, account)
}

func (c *Environment) Delete(ctx context.Context, account *orgv1.Account) error {
	c.Logger.Log("msg", "Environment.Delete()")
	return c.Client.Account.Delete(ctx, account)
}

func (c *Environment) Get(ctx context.Context, account *orgv1.Account) (*orgv1.Account, error) {
	c.Logger.Log("msg", "Environment.Get()")
	return c.Client.Account.Get(ctx, account)
}

func (c *Environment) List(ctx context.Context, account *orgv1.Account) ([]*orgv1.Account, error) {
	c.Logger.Log("msg", "Environment.List()")
	return c.Client.Account.List(ctx, account)
}
