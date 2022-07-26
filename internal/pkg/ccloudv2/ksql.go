package ccloudv2

import (
	"context"

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

func (c *Client) ListKsqlClusters(environmentId string) (ksql.KsqldbcmV2ClusterList, error) {
	clusters, _, err := c.KsqlClient.ClustersKsqldbcmV2Api.ListKsqldbcmV2Clusters(c.ksqlApiContext()).Environment(environmentId).Execute()
	return clusters, err
}

func (c *Client) DeleteKsqlCluster(clusterId, environmentId string) error {
	_, err := c.KsqlClient.ClustersKsqldbcmV2Api.DeleteKsqldbcmV2Cluster(c.ksqlApiContext(), clusterId).Environment(environmentId).Execute()
	return err
}

func (c *Client) DescribeKsqlCluster(clusterId, environmentId string) (ksql.KsqldbcmV2Cluster, error) {
	cluster, _, err := c.KsqlClient.ClustersKsqldbcmV2Api.GetKsqldbcmV2Cluster(c.ksqlApiContext(), clusterId).Environment(environmentId).Execute()
	return cluster, err
}
