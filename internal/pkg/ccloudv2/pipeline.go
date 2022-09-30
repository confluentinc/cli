package ccloudv2

import (
	"context"
	sdv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newSdClient(url, userAgent string, unsafeTrace bool) *sdv1.APIClient {
	cfg := sdv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = sdv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return sdv1.NewAPIClient(cfg)
}

func (c *Client) sdApiContext() context.Context {
	return context.WithValue(context.Background(), sdv1.ContextAccessToken, c.AuthToken)
}

// sd pipeline api calls

func (c *Client) ListPipelines(envId string, clusterId string) ([]sdv1.SdV1Pipeline, error) {
	var list []sdv1.SdV1Pipeline

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListPipelines(envId, clusterId, pageToken)
		if err != nil {
			return nil, err
		}
		list = append(list, page.GetData()...)

		pageToken, done, err = extractSdNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) executeListPipelines(envId, clusterId, pageToken string) (sdv1.SdV1PipelineList, error) {
	req := c.SdClient.PipelinesSdV1Api.ListSdV1Pipelines(c.sdApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	req = req.Environment(envId).SpecKafkaCluster(clusterId)

	resp, httpResp, err := c.SdClient.PipelinesSdV1Api.ListSdV1PipelinesExecute(req)
	return resp, errors.CatchStreamDesignerError(err, httpResp)
}

func (c *Client) CreatePipeline(envId, clusterId, name, description, ksqlId, srClusterId string) (sdv1.SdV1Pipeline, error) {
	createPipeline := sdv1.SdV1Pipeline{
		Spec: &sdv1.SdV1PipelineSpec{
			DisplayName:             sdv1.PtrString(name),
			Description:             sdv1.PtrString(description),
			Environment:             &sdv1.ObjectReference{Id: envId},
			KafkaCluster:            &sdv1.ObjectReference{Id: clusterId},
			KsqlCluster:             &sdv1.ObjectReference{Id: ksqlId},
			StreamGovernanceCluster: &sdv1.ObjectReference{Id: srClusterId},
		},
	}

	req := c.SdClient.PipelinesSdV1Api.CreateSdV1Pipeline(c.sdApiContext()).SdV1Pipeline(createPipeline)
	resp, httpResp, err := c.SdClient.PipelinesSdV1Api.CreateSdV1PipelineExecute(req)
	return resp, errors.CatchStreamDesignerError(err, httpResp)
}

func (c *Client) DeleteSdPipeline(envId, clusterId, id string) error {
	req := c.SdClient.PipelinesSdV1Api.DeleteSdV1Pipeline(c.sdApiContext(), id)
	req = req.Environment(envId).SpecKafkaCluster(clusterId)

	httpResp, err := c.SdClient.PipelinesSdV1Api.DeleteSdV1PipelineExecute(req)
	return errors.CatchStreamDesignerError(err, httpResp)
}

func (c *Client) GetSdPipeline(envId, clusterId, id string) (sdv1.SdV1Pipeline, error) {
	req := c.SdClient.PipelinesSdV1Api.GetSdV1Pipeline(c.sdApiContext(), id).Environment(envId).SpecKafkaCluster(clusterId)

	resp, httpResp, err := c.SdClient.PipelinesSdV1Api.GetSdV1PipelineExecute(req)
	return resp, errors.CatchStreamDesignerError(err, httpResp)
}

func (c *Client) UpdateSdPipeline(envId string, clusterId string, id string, update sdv1.SdV1PipelineUpdate) (sdv1.SdV1Pipeline, error) {
	update.Spec.SetEnvironment(sdv1.ObjectReference{Id: envId})
	update.Spec.SetKafkaCluster(sdv1.ObjectReference{Id: clusterId})

	req := c.SdClient.PipelinesSdV1Api.UpdateSdV1Pipeline(c.sdApiContext(), id).SdV1PipelineUpdate(update)

	resp, httpResp, err := c.SdClient.PipelinesSdV1Api.UpdateSdV1PipelineExecute(req)
	return resp, errors.CatchStreamDesignerError(err, httpResp)
}

func extractSdNextPageToken(nextPageUrlStringNullable sdv1.NullableString) (string, bool, error) {
	if !nextPageUrlStringNullable.IsSet() {
		return "", true, nil
	}
	nextPageUrlString := *nextPageUrlStringNullable.Get()
	pageToken, err := extractPageToken(nextPageUrlString)
	return pageToken, false, err
}
