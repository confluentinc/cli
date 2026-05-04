package ccloudv2

import (
	"context"
	"net/http"

	rtcev1 "github.com/confluentinc/ccloud-sdk-go-v2/rtce/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// ===== API group client bootstrap =====

func newRtceClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *rtcev1.APIClient {
	cfg := rtcev1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = rtcev1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return rtcev1.NewAPIClient(cfg)
}

func (c *Client) rtceApiContext() context.Context {
	return context.WithValue(context.Background(), rtcev1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

// ===== rtce topics API calls =====

func (c *Client) CreateRtceTopic(req rtcev1.RtceV1RtceTopic) (rtcev1.RtceV1RtceTopic, *http.Response, error) {
	createReq := c.RtceClient.RtceTopicsRtceV1Api.
		CreateRtceV1RtceTopic(c.rtceApiContext()).
		RtceV1RtceTopic(req)
	return createReq.Execute()
}

func (c *Client) GetRtceTopic(topicName string, environment string, specKafkaCluster string) (rtcev1.RtceV1RtceTopic, *http.Response, error) {
	getReq := c.RtceClient.RtceTopicsRtceV1Api.
		GetRtceV1RtceTopic(c.rtceApiContext(), topicName)
	getReq = getReq.Environment(environment)
	getReq = getReq.SpecKafkaCluster(specKafkaCluster)
	return getReq.Execute()
}

func (c *Client) UpdateRtceTopic(topicName string, update rtcev1.RtceV1RtceTopicUpdate) (rtcev1.RtceV1RtceTopic, *http.Response, error) {
	updateReq := c.RtceClient.RtceTopicsRtceV1Api.
		UpdateRtceV1RtceTopic(c.rtceApiContext(), topicName).
		RtceV1RtceTopicUpdate(update)
	return updateReq.Execute()
}

func (c *Client) DeleteRtceTopic(topicName string, environment string, specKafkaCluster string) error {
	deleteReq := c.RtceClient.RtceTopicsRtceV1Api.
		DeleteRtceV1RtceTopic(c.rtceApiContext(), topicName)
	deleteReq = deleteReq.Environment(environment)
	deleteReq = deleteReq.SpecKafkaCluster(specKafkaCluster)
	httpResp, err := deleteReq.Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListRtceTopics(specCloud string, specRegion string, environment string, specKafkaCluster string) ([]rtcev1.RtceV1RtceTopic, error) {
	var list []rtcev1.RtceV1RtceTopic

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListRtceTopics(specCloud, specRegion, environment, specKafkaCluster, pageToken)
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

func (c *Client) executeListRtceTopics(specCloud string, specRegion string, environment string, specKafkaCluster string, pageToken string) (rtcev1.RtceV1RtceTopicList, *http.Response, error) {
	req := c.RtceClient.RtceTopicsRtceV1Api.
		ListRtceV1RtceTopics(c.rtceApiContext()).
		Environment(environment).
		SpecKafkaCluster(specKafkaCluster).
		PageSize(ccloudV2ListPageSize)
	if specCloud != "" {
		req = req.SpecCloud(specCloud)
	}
	if specRegion != "" {
		req = req.SpecRegion(specRegion)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

// ===== rtce regions API calls =====

func (c *Client) ListRtceRegions(cloud string, region string) ([]rtcev1.RtceV1Region, error) {
	var list []rtcev1.RtceV1Region

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListRegions(cloud, region, pageToken)
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

func (c *Client) executeListRegions(cloud string, region string, pageToken string) (rtcev1.RtceV1RegionList, *http.Response, error) {
	req := c.RtceClient.RegionsRtceV1Api.
		ListRtceV1Regions(c.rtceApiContext()).
		PageSize(ccloudV2ListPageSize)
	if cloud != "" {
		req = req.Cloud(cloud)
	}
	if region != "" {
		req = req.Region(region)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
