package ccloudv2

import (
	"context"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newStreamDesignerClient(url, userAgent string, unsafeTrace bool) *streamdesignerv1.APIClient {
	cfg := streamdesignerv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = streamdesignerv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return streamdesignerv1.NewAPIClient(cfg)
}

func (c *Client) sdApiContext() context.Context {
	return context.WithValue(context.Background(), streamdesignerv1.ContextAccessToken, c.AuthToken)
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
	req := c.StreamDesignerClient.PipelinesSdV1Api.ListSdV1Pipelines(c.sdApiContext()).PageSize(ccloudV2ListPageSize).Environment(envId).SpecKafkaCluster(clusterId)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := c.StreamDesignerClient.PipelinesSdV1Api.ListSdV1PipelinesExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreatePipeline(createPipeline streamdesignerv1.SdV1Pipeline) (streamdesignerv1.SdV1Pipeline, error) {
	req := c.StreamDesignerClient.PipelinesSdV1Api.CreateSdV1Pipeline(c.sdApiContext()).SdV1Pipeline(createPipeline)
	resp, httpResp, err := c.StreamDesignerClient.PipelinesSdV1Api.CreateSdV1PipelineExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteSdPipeline(envId, clusterId, id string) error {
	req := c.StreamDesignerClient.PipelinesSdV1Api.DeleteSdV1Pipeline(c.sdApiContext(), id)
	req = req.Environment(envId).SpecKafkaCluster(clusterId)

	httpResp, err := c.StreamDesignerClient.PipelinesSdV1Api.DeleteSdV1PipelineExecute(req)
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetSdPipeline(envId, clusterId, id string) (streamdesignerv1.SdV1Pipeline, error) {
	req := c.StreamDesignerClient.PipelinesSdV1Api.GetSdV1Pipeline(c.sdApiContext(), id).Environment(envId).SpecKafkaCluster(clusterId)

	resp, httpResp, err := c.StreamDesignerClient.PipelinesSdV1Api.GetSdV1PipelineExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateSdPipeline(envId, clusterId, id string, update streamdesignerv1.SdV1Pipeline) (streamdesignerv1.SdV1Pipeline, error) {
	update.Spec.SetEnvironment(streamdesignerv1.ObjectReference{Id: envId})
	update.Spec.SetKafkaCluster(streamdesignerv1.ObjectReference{Id: clusterId})

	req := c.StreamDesignerClient.PipelinesSdV1Api.UpdateSdV1Pipeline(c.sdApiContext(), id).SdV1Pipeline(update)

	resp, httpResp, err := c.StreamDesignerClient.PipelinesSdV1Api.UpdateSdV1PipelineExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
