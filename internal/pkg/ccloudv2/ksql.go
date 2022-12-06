package ccloudv2

import (
	"context"

	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newKsqlClient(url, userAgent string, unsafeTrace bool) *ksqlv2.APIClient {
	cfg := ksqlv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = ksqlv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return ksqlv2.NewAPIClient(cfg)
}

func (c *Client) ksqlApiContext() context.Context {
	return context.WithValue(context.Background(), ksqlv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) ListKsqlClusters(environmentId string) (ksqlv2.KsqldbcmV2ClusterList, error) {
	clusters, resp, err := c.KsqlClient.ClustersKsqldbcmV2Api.ListKsqldbcmV2Clusters(c.ksqlApiContext()).Environment(environmentId).Execute()
	return clusters, errors.CatchCCloudV2Error(err, resp)
}

func (c *Client) DeleteKsqlCluster(clusterId, environmentId string) error {
	resp, err := c.KsqlClient.ClustersKsqldbcmV2Api.DeleteKsqldbcmV2Cluster(c.ksqlApiContext(), clusterId).Environment(environmentId).Execute()
	return errors.CatchCCloudV2Error(err, resp)
}

func (c *Client) DescribeKsqlCluster(clusterId, environmentId string) (ksqlv2.KsqldbcmV2Cluster, error) {
	cluster, resp, err := c.KsqlClient.ClustersKsqldbcmV2Api.GetKsqldbcmV2Cluster(c.ksqlApiContext(), clusterId).Environment(environmentId).Execute()
	return cluster, errors.CatchCCloudV2Error(err, resp)
}

func (c *Client) CreateKsqlCluster(displayName, environmentId, kafkaClusterId, credentialIdentity string, csus int32, useDetailedProcessingLog bool) (ksqlv2.KsqldbcmV2Cluster, error) {
	cluster := ksqlv2.KsqldbcmV2Cluster{
		Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
			DisplayName:              &displayName,
			UseDetailedProcessingLog: &useDetailedProcessingLog,
			Csu:                      &csus,
			KafkaCluster:       &ksqlv2.ObjectReference{Id: kafkaClusterId, Related: "-", ResourceName: "-"},
			CredentialIdentity: &ksqlv2.ObjectReference{Id: credentialIdentity, Related: "-", ResourceName: "-"},
			Environment:        &ksqlv2.ObjectReference{Id: environmentId, Related: "-", ResourceName: "-"},
		},
	}
	created, resp, err := c.KsqlClient.ClustersKsqldbcmV2Api.CreateKsqldbcmV2Cluster(c.ksqlApiContext()).KsqldbcmV2Cluster(cluster).Execute()
	return created, errors.CatchCCloudV2Error(err, resp)
}
