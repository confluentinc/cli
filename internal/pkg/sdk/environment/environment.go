package environment

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
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
	ret, err := c.Client.Account.Create(ctx, account)
	return ret, errors.ConvertAPIError(err)
}

func (c *Environment) Update(ctx context.Context, account *orgv1.Account) error {
	c.Logger.Log("msg", "Environment.Update()")
	err := c.Client.Account.Update(ctx, account)
	return errors.ConvertAPIError(err)
}

func (c *Environment) Delete(ctx context.Context, account *orgv1.Account) error {
	c.Logger.Log("msg", "Environment.Delete()")
	err := c.Client.Account.Delete(ctx, account)
	return errors.ConvertAPIError(err)
}

func (c *Environment) Get(ctx context.Context, account *orgv1.Account) (*orgv1.Account, error) {
	c.Logger.Log("msg", "Environment.Get()")
	ret, err := c.Client.Account.Get(ctx, account)
	return ret, errors.ConvertAPIError(err)
}

func (c *Environment) List(ctx context.Context, account *orgv1.Account) ([]*orgv1.Account, error) {
	c.Logger.Log("msg", "Environment.List()")
	ret, err := c.Client.Account.List(ctx, account)
	return ret, errors.ConvertAPIError(err)
}
