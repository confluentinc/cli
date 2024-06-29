package ccloudv2

import (
	"context"
	"net/http"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newStreamDesignerClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *streamdesignerv1.APIClient {
	cfg := streamdesignerv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = streamdesignerv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return streamdesignerv1.NewAPIClient(cfg)
}

func (c *Client) streamDesignerApiContext() context.Context {
	return context.WithValue(context.Background(), streamdesignerv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) ListPipelines(envId, clusterId string) ([]streamdesignerv1.SdV1Pipeline, error) {
	var list []streamdesignerv1.SdV1Pipeline

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListPipelines(envId, clusterId, pageToken)
		if err != nil {
			return nil, err
		}
		list = append(list, page.GetData()...)

		pageToken, done, err = extractNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) executeListPipelines(envId, clusterId, pageToken string) (streamdesignerv1.SdV1PipelineList, error) {
	req := c.StreamDesignerClient.PipelinesSdV1Api.ListSdV1Pipelines(c.streamDesignerApiContext()).PageSize(ccloudV2ListPageSize).Environment(envId).SpecKafkaCluster(clusterId)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreatePipeline(createPipeline streamdesignerv1.SdV1Pipeline) (streamdesignerv1.SdV1Pipeline, error) {
	resp, httpResp, err := c.StreamDesignerClient.PipelinesSdV1Api.CreateSdV1Pipeline(c.streamDesignerApiContext()).SdV1Pipeline(createPipeline).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteSdPipeline(envId, clusterId, id string) error {
	httpResp, err := c.StreamDesignerClient.PipelinesSdV1Api.DeleteSdV1Pipeline(c.streamDesignerApiContext(), id).Environment(envId).SpecKafkaCluster(clusterId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetSdPipeline(envId, clusterId, id string) (streamdesignerv1.SdV1Pipeline, error) {
	resp, httpResp, err := c.StreamDesignerClient.PipelinesSdV1Api.GetSdV1Pipeline(c.streamDesignerApiContext(), id).Environment(envId).SpecKafkaCluster(clusterId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateSdPipeline(id string, pipeline streamdesignerv1.SdV1Pipeline) (streamdesignerv1.SdV1Pipeline, error) {
	resp, httpResp, err := c.StreamDesignerClient.PipelinesSdV1Api.UpdateSdV1Pipeline(c.streamDesignerApiContext(), id).SdV1Pipeline(pipeline).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
