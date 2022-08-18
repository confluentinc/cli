package ccloudv2

import (
	"context"
	"net/http"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
)

func newCliClient(url, userAgent string, unsafeTrace bool) *cliv1.APIClient {
	// We do not use a retryable HTTP client so the CLI does not hang if there is a problem with the usage service.

	cfg := cliv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.Servers = cliv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return cliv1.NewAPIClient(cfg)
}

func (c *Client) cliApiContext() context.Context {
	return context.WithValue(context.Background(), cliv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateCliUsage(usage cliv1.CliV1Usage) (*http.Response, error) {
	req := c.CliClient.UsagesCliV1Api.CreateCliV1Usage(c.cliApiContext()).CliV1Usage(usage)
	return c.CliClient.UsagesCliV1Api.CreateCliV1UsageExecute(req)
}
