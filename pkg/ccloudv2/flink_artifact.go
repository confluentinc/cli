package ccloudv2

import (
	"context"
	"net/http"

	flinkartifactv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-artifact/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newFlinkArtifactClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *flinkartifactv1.APIClient {
	cfg := flinkartifactv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = flinkartifactv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return flinkartifactv1.NewAPIClient(cfg)
}

func (c *Client) flinkArtifactApiContext() context.Context {
	return context.WithValue(context.Background(), flinkartifactv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) GetFlinkPresignedUrl(request flinkartifactv1.ArtifactV1PresignedUrlRequest) (flinkartifactv1.ArtifactV1PresignedUrl, error) {
	resp, httpResp, err := c.FlinkArtifactClient.PresignedUrlsArtifactV1Api.PresignedUploadUrlArtifactV1PresignedUrl(c.flinkArtifactApiContext()).ArtifactV1PresignedUrlRequest(request).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateFlinkArtifact(createFlinkArtifactRequest flinkartifactv1.InlineObject) (flinkartifactv1.ArtifactV1FlinkArtifact, error) {
	resp, httpResp, err := c.FlinkArtifactClient.FlinkArtifactsArtifactV1Api.CreateArtifactV1FlinkArtifact(c.flinkArtifactApiContext()).
		Cloud(createFlinkArtifactRequest.Cloud).Region(createFlinkArtifactRequest.Region).InlineObject(createFlinkArtifactRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListFlinkArtifacts(cloud string, region string, env string) ([]flinkartifactv1.ArtifactV1FlinkArtifact, error) {
	var list []flinkartifactv1.ArtifactV1FlinkArtifact

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListArtifacts(pageToken, cloud, region, env)
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

func (c *Client) DescribeFlinkArtifact(cloud string, region string, id string) (flinkartifactv1.ArtifactV1FlinkArtifact, error) {
	resp, httpResp, err := c.FlinkArtifactClient.FlinkArtifactsArtifactV1Api.GetArtifactV1FlinkArtifact(c.flinkArtifactApiContext(), id).Cloud(cloud).Region(region).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteFlinkArtifact(cloud string, region string, id string) error {
	httpResp, err := c.FlinkArtifactClient.FlinkArtifactsArtifactV1Api.DeleteArtifactV1FlinkArtifact(c.flinkArtifactApiContext(), id).Cloud(cloud).Region(region).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateFlinkArtifact(id string, updateRequest flinkartifactv1.ArtifactV1FlinkArtifactUpdate) (flinkartifactv1.ArtifactV1FlinkArtifact, error) {
	resp, httpResp, err := c.FlinkArtifactClient.FlinkArtifactsArtifactV1Api.UpdateArtifactV1FlinkArtifact(c.flinkArtifactApiContext(), id).ArtifactV1FlinkArtifactUpdate(updateRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) executeListArtifacts(pageToken, cloud string, region string, env string) (flinkartifactv1.ArtifactV1FlinkArtifactList, *http.Response, error) {
	req := c.FlinkArtifactClient.FlinkArtifactsArtifactV1Api.ListArtifactV1FlinkArtifacts(c.flinkArtifactApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	if cloud != "" {
		req = req.Cloud(cloud)
	}
	if region != "" {
		req = req.Region(region)
	}
	if env != "" {
		req = req.Environment(env)
	}
	return req.Execute()
}
