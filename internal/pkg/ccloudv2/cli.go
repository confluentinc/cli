package ccloudv2

import (
	"context"
	"net/http"

	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"

	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newCliClient(baseURL, userAgent string, isTest bool) *cliv1.APIClient {
	cfg := cliv1.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = cliv1.ServerConfigurations{{URL: getServerUrl(baseURL, isTest), Description: "Confluent Cloud CLI"}}
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
