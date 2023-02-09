package ccloudv2

import (
	"context"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newCliClient(url, userAgent string, unsafeTrace bool) *cliv1.APIClient {
	// We do not use a retryable HTTP client so the CLI does not hang if there is a problem with the usage service.

	cfg := cliv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.Servers = cliv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return cliv1.NewAPIClient(cfg)
}

func (c *Client) cliApiContext(ctx context.Context) context.Context {
	return context.WithValue(context.Background(), cliv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateCliUsage(ctx context.Context, usage cliv1.CliV1Usage) error {
	req := c.CliClient.UsagesCliV1Api.CreateCliV1Usage(c.cliApiContext(ctx)).CliV1Usage(usage)
	r, err := c.CliClient.UsagesCliV1Api.CreateCliV1UsageExecute(req)
	return errors.CatchCCloudV2Error(err, r)
}
