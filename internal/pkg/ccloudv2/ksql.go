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

func (c *Client) CreateKsqlCluster(displayName, environmentId, kafkaClusterId, credentialIdentity string, csus int32, logExcludeRows bool) (ksql.KsqldbcmV2Cluster, *_nethttp.Response, error) {
	cluster := ksql.KsqldbcmV2Cluster{
		Spec: &ksql.KsqldbcmV2ClusterSpec{
			DisplayName:        &displayName,
			Csu:                &csus,
			KafkaCluster:       &ksql.ObjectReference{Id: kafkaClusterId},
			CredentialIdentity: &ksql.ObjectReference{Id: credentialIdentity},
			Environment:        &ksql.ObjectReference{Id: environmentId},
		},
	}
	return c.KsqlClient.ClustersKsqldbcmV2Api.CreateKsqldbcmV2Cluster(c.ksqlApiContext()).KsqldbcmV2Cluster(cluster).Execute()
}

func (c *Client) DescribeKsqlCluster(ksqlClusterId string) (ksql.KsqldbcmV2Cluster, *_nethttp.Response, error) {
	return c.KsqlClient.ClustersKsqldbcmV2Api.GetKsqldbcmV2Cluster(c.ksqlApiContext(), ksqlClusterId).Execute()
}