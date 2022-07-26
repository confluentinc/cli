package ccloudv2

import (
	"context"
	_nethttp "net/http"

	ksql "github.com/confluentinc/ccloud-sdk-go-v2-internal/ksql/v2"

	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newKsqlClient(baseURL, userAgent string, isTest bool) *ksql.APIClient {
	cfg := ksql.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = ksql.ServerConfigurations{{URL: getServerUrl(baseURL, isTest), Description: "Confluent Cloud KSQL"}}
	cfg.UserAgent = userAgent

	return ksql.NewAPIClient(cfg)
}


func (c *Client) ksqlApiContext() context.Context {
	return context.WithValue(context.Background(), ksql.ContextAccessToken, c.AuthToken)
}

func (c *Client) ListKsqlClusters(environmentId string) (ksql.KsqldbcmV2ClusterList, *_nethttp.Response, error) {
	return c.KsqlClient.ClustersKsqldbcmV2Api.ListKsqldbcmV2Clusters(c.ksqlApiContext()).Environment(environmentId).Execute()
}
