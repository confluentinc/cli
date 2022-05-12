package ccloudv2

import (
	"context"
	sgv2 "github.com/confluentinc/ccloud-sdk-go-v2/stream-governance/v2"
	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newStreamGovernanceClient(baseURL, userAgent string, isTest bool) *sgv2.APIClient {
	streamGovernanceServer := getServerUrl(baseURL, isTest)
	cfg := sgv2.NewConfiguration()
	cfg.Servers = sgv2.ServerConfigurations{
		{URL: streamGovernanceServer, Description: "Confluent Cloud Stream-Governance"},
	}
	cfg.UserAgent = userAgent
	cfg.Debug = plog.CliLogger.GetLevel() >= plog.DEBUG
	return sgv2.NewAPIClient(cfg)
}

func (c *Client) StreamGovernanceApiContext() context.Context {
	return context.WithValue(context.Background(), sgv2.ContextAccessToken, c.AuthToken)
}
