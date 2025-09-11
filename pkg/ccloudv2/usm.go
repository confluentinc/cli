package ccloudv2

import (
	"context"
	"net/http"

	usmv1 "github.com/confluentinc/ccloud-sdk-go-v2/usm/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newUsmClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *usmv1.APIClient {
	cfg := usmv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = usmv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return usmv1.NewAPIClient(cfg)
}

func (c *Client) usmApiContext() context.Context {
	return context.WithValue(context.Background(), usmv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) CreateUsmKafkaCluster(cluster usmv1.UsmV1KafkaCluster) (usmv1.UsmV1KafkaCluster, error) {
	resp, httpResp, err := c.UsmClient.KafkaClustersUsmV1Api.CreateUsmV1KafkaCluster(c.usmApiContext()).UsmV1KafkaCluster(cluster).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteUsmKafkaCluster(id, environment string) error {
	httpResp, err := c.UsmClient.KafkaClustersUsmV1Api.DeleteUsmV1KafkaCluster(c.usmApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetUsmKafkaCluster(id, environment string) (usmv1.UsmV1KafkaCluster, error) {
	resp, httpResp, err := c.UsmClient.KafkaClustersUsmV1Api.GetUsmV1KafkaCluster(c.usmApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListUsmKafkaClusters(environment string) ([]usmv1.UsmV1KafkaCluster, error) {
	var list []usmv1.UsmV1KafkaCluster

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListUsmKafkaClusters(environment, pageToken)
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

func (c *Client) executeListUsmKafkaClusters(environment, pageToken string) (usmv1.UsmV1KafkaClusterList, *http.Response, error) {
	req := c.UsmClient.KafkaClustersUsmV1Api.ListUsmV1KafkaClusters(c.usmApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) CreateUsmConnectCluster(cluster usmv1.UsmV1ConnectCluster) (usmv1.UsmV1ConnectCluster, error) {
	resp, httpResp, err := c.UsmClient.ConnectClustersUsmV1Api.CreateUsmV1ConnectCluster(c.usmApiContext()).UsmV1ConnectCluster(cluster).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteUsmConnectCluster(id, environment string) error {
	httpResp, err := c.UsmClient.ConnectClustersUsmV1Api.DeleteUsmV1ConnectCluster(c.usmApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetUsmConnectCluster(id, environment string) (usmv1.UsmV1ConnectCluster, error) {
	resp, httpResp, err := c.UsmClient.ConnectClustersUsmV1Api.GetUsmV1ConnectCluster(c.usmApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListUsmConnectClusters(environment string) ([]usmv1.UsmV1ConnectCluster, error) {
	var list []usmv1.UsmV1ConnectCluster

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListUsmConnectClusters(environment, pageToken)
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

func (c *Client) executeListUsmConnectClusters(environment, pageToken string) (usmv1.UsmV1ConnectClusterList, *http.Response, error) {
	req := c.UsmClient.ConnectClustersUsmV1Api.ListUsmV1ConnectClusters(c.usmApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
