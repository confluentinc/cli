package ccloudv2

import (
	"context"
	"fmt"
	"net/http"

	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newKsqlClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *ksqlv2.APIClient {
	cfg := ksqlv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = ksqlv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return ksqlv2.NewAPIClient(cfg)
}

func (c *Client) ksqlApiContext() context.Context {
	return context.WithValue(context.Background(), ksqlv2.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) ListKsqlClusters(environmentId string) ([]ksqlv2.KsqldbcmV2Cluster, error) {
	var list []ksqlv2.KsqldbcmV2Cluster

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListKsqlClusters(pageToken, environmentId)
		if err != nil {
			return nil, errors.CatchCCloudV2Error(err, httpResp)
		}
		list = append(list, page.GetData()...)

		pageToken, done, err = extractNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) executeListKsqlClusters(pageToken, environmentId string) (ksqlv2.KsqldbcmV2ClusterList, *http.Response, error) {
	req := c.KsqlClient.ClustersKsqldbcmV2Api.ListKsqldbcmV2Clusters(c.ksqlApiContext()).Environment(environmentId).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) DeleteKsqlCluster(clusterId, environmentId string) error {
	httpResp, err := c.KsqlClient.ClustersKsqldbcmV2Api.DeleteKsqldbcmV2Cluster(c.ksqlApiContext(), clusterId).Environment(environmentId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeKsqlCluster(clusterId, environmentId string) (ksqlv2.KsqldbcmV2Cluster, error) {
	res, httpResp, err := c.KsqlClient.ClustersKsqldbcmV2Api.GetKsqldbcmV2Cluster(c.ksqlApiContext(), clusterId).Environment(environmentId).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateKsqlCluster(displayName, environmentId, kafkaClusterId, credentialIdentity string, csus int32, useDetailedProcessingLog bool) (ksqlv2.KsqldbcmV2Cluster, error) {
	cluster := ksqlv2.KsqldbcmV2Cluster{Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
		DisplayName:              &displayName,
		UseDetailedProcessingLog: &useDetailedProcessingLog,
		Csu:                      &csus,
		KafkaCluster:             &ksqlv2.ObjectReference{Id: kafkaClusterId, Related: "-", ResourceName: "-"},
		CredentialIdentity:       &ksqlv2.ObjectReference{Id: credentialIdentity, Related: "-", ResourceName: "-"},
		Environment:              &ksqlv2.ObjectReference{Id: environmentId, Related: "-", ResourceName: "-"},
	}}
	res, httpResp, err := c.KsqlClient.ClustersKsqldbcmV2Api.CreateKsqldbcmV2Cluster(c.ksqlApiContext()).KsqldbcmV2Cluster(cluster).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

// UpdateKsqlCluster issues PATCH /ksqldbcm/v2/clusters/{id} with {"spec":{"csu": N}}
// to trigger a self-serve cluster resize.
//
// The PATCH operation is not yet available in ccloud-sdk-go-v2/ksql/v2 — it is
// being added in cc-api PR #2507 (KSQL-14844), after which the SDK needs to be
// regenerated and the ksql module dependency in go.mod bumped. Until that
// lands, calling this method returns a clear, customer-safe error rather than
// an HTTP failure. See KSQL-14849 for the work item.
//
// Wiring instructions once the SDK is regenerated:
//
//	cluster := ksqlv2.KsqldbcmV2Cluster{Spec: &ksqlv2.KsqldbcmV2ClusterSpec{Csu: &csu}}
//	res, httpResp, err := c.KsqlClient.ClustersKsqldbcmV2Api.
//	    UpdateKsqldbcmV2Cluster(c.ksqlApiContext(), id).
//	    KsqldbcmV2ClusterUpdate(cluster).Environment(environmentId).Execute()
//	return res, errors.CatchCCloudV2Error(err, httpResp)
func (c *Client) UpdateKsqlCluster(id, environmentId string, csu int32) (ksqlv2.KsqldbcmV2Cluster, error) {
	_ = id
	_ = environmentId
	_ = csu
	return ksqlv2.KsqldbcmV2Cluster{}, fmt.Errorf(
		"ksqlDB cluster update is not yet available in this CLI build. " +
			"Please upgrade to a newer version of the Confluent CLI and retry.")
}
