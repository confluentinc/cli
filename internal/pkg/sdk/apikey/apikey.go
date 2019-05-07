package apikey

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

// Compile-time check for Interface adherence
var _ ccloud.APIKey = (*APIKey)(nil)

type APIKey struct {
	Logger *log.Logger
	Client *ccloud.Client
}

func New(client *ccloud.Client, logger *log.Logger) *APIKey {
	return &APIKey{Client: client, Logger: logger}
}

func (c *APIKey) Create(ctx context.Context, key *authv1.ApiKey) (*authv1.ApiKey, error) {
	c.Logger.Log("msg", "apiKey.Create()")
	ret, err := c.Client.APIKey.Create(ctx, key)
	return ret, errors.ConvertAPIError(err)
}

func (c *APIKey) Update(ctx context.Context, key *authv1.ApiKey) error {
	c.Logger.Log("msg", "apiKey.Update()")
	err := c.Client.APIKey.Update(ctx, key)
	return errors.ConvertAPIError(err)
}

func (c *APIKey) Delete(ctx context.Context, key *authv1.ApiKey) error {
	c.Logger.Log("msg", "apiKey.Delete()")
	err := c.Client.APIKey.Delete(ctx, key)
	return errors.ConvertAPIError(err)
}

func (c *APIKey) List(ctx context.Context, key *authv1.ApiKey) ([]*authv1.ApiKey, error) {
	c.Logger.Log("msg", "apiKey.List()")
	ret, err := c.Client.APIKey.List(ctx, key)
	return ret, errors.ConvertAPIError(err)
}

func (c *APIKey) Get(ctx context.Context, key *authv1.ApiKey) (*authv1.ApiKey, error) {
	c.Logger.Log("msg", "apiKey.Get()")
	ret, err := c.Client.APIKey.Get(ctx, key)
	return ret, errors.ConvertAPIError(err)
}
