package ccloudv2

import (
	"context"

	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newKsqlClient(baseURL, userAgent string, isTest, unsafeTrace bool) *ksqlv2.APIClient {
	cfg := ksqlv2.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = ksqlv2.ServerConfigurations{{URL: getServerUrl(baseURL, isTest)}}
	cfg.UserAgent = userAgent

	return ksqlv2.NewAPIClient(cfg)
}

func (c *Client) ksqlApiContext() context.Context {
	return context.WithValue(context.Background(), ksqlv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) ListKsqlClusters(environmentId string) (ksqlv2.KsqldbcmV2ClusterList, error) {
	clusters, resp, err := c.KsqlClient.ClustersKsqldbcmV2Api.ListKsqldbcmV2Clusters(c.ksqlApiContext()).Environment(environmentId).Execute()
	return clusters, errors.CatchV2ErrorDetailWithResponse(err, resp)
}

func (c *Client) DeleteKsqlCluster(clusterId, environmentId string) error {
	resp, err := c.KsqlClient.ClustersKsqldbcmV2Api.DeleteKsqldbcmV2Cluster(c.ksqlApiContext(), clusterId).Environment(environmentId).Execute()
	return errors.CatchV2ErrorDetailWithResponse(err, resp)
}

func (c *Client) DescribeKsqlCluster(clusterId, environmentId string) (ksqlv2.KsqldbcmV2Cluster, error) {
	cluster, resp, err := c.KsqlClient.ClustersKsqldbcmV2Api.GetKsqldbcmV2Cluster(c.ksqlApiContext(), clusterId).Environment(environmentId).Execute()
	return cluster, errors.CatchV2ErrorDetailWithResponse(err, resp)
}

func (c *Client) CreateKsqlCluster(displayName, environmentId, kafkaClusterId, credentialIdentity string, csus int32, useDetailedProcessingLog bool) (ksqlv2.KsqldbcmV2Cluster, error) {
	cluster := ksqlv2.KsqldbcmV2Cluster{
		Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
			DisplayName:              &displayName,
			UseDetailedProcessingLog: &useDetailedProcessingLog,
			Csu:                      &csus,
			// TODO: remove cat (See https://github.com/confluentinc/cli/pull/1371/files#r949420697)
			KafkaCluster:       &ksqlv2.ObjectReference{Id: kafkaClusterId, Related: "cat", ResourceName: "cat"},
			CredentialIdentity: &ksqlv2.ObjectReference{Id: credentialIdentity, Related: "cat", ResourceName: "cat"},
			Environment:        &ksqlv2.ObjectReference{Id: environmentId, Related: "cat", ResourceName: "cat"},
		},
	}
	created, resp, err := c.KsqlClient.ClustersKsqldbcmV2Api.CreateKsqldbcmV2Cluster(c.ksqlApiContext()).KsqldbcmV2Cluster(cluster).Execute()
	return created, errors.CatchV2ErrorDetailWithResponse(err, resp)
}
