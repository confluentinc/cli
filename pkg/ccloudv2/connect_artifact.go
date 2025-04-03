package ccloudv2

import (
	"context"
	"net/http"

	camv1 "github.com/confluentinc/ccloud-sdk-go-v2/cam/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newConnectArtifactClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *camv1.APIClient {
	cfg := camv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = camv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return camv1.NewAPIClient(cfg)
}

func (c *Client) connectArtifactApiContext() context.Context {
	return context.WithValue(context.Background(), camv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) GetArtifactPresignedUrl(request camv1.CamV1PresignedUrlRequest) (camv1.CamV1PresignedUrl, error) {
	resp, httpResp, err := c.ConnectArtifactClient.PresignedUrlsCamV1Api.PresignedUploadUrlCamV1PresignedUrl(c.connectArtifactApiContext()).CamV1PresignedUrlRequest(request).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateConnectArtifact(createArtifactRequest camv1.CamV1ConnectArtifact) (camv1.CamV1ConnectArtifact, error) {
	resp, httpResp, err := c.ConnectArtifactClient.ConnectArtifactsCamV1Api.CreateCamV1ConnectArtifact(c.connectArtifactApiContext()).CamV1ConnectArtifact(createArtifactRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListConnectArtifacts(cloud, region, env string) ([]camv1.CamV1ConnectArtifact, error) {
	var list []camv1.CamV1ConnectArtifact

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListConnectArtifacts(pageToken, cloud, region, env)
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

func (c *Client) DescribeConnectArtifact(cloud, region, environment, id string) (camv1.CamV1ConnectArtifact, error) {
	resp, httpResp, err := c.ConnectArtifactClient.ConnectArtifactsCamV1Api.GetCamV1ConnectArtifact(c.connectArtifactApiContext(), id).SpecCloud(cloud).SpecRegion(region).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteConnectArtifact(cloud, region, environment, id string) error {
	httpResp, err := c.ConnectArtifactClient.ConnectArtifactsCamV1Api.DeleteCamV1ConnectArtifact(c.connectArtifactApiContext(), id).SpecCloud(cloud).SpecRegion(region).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

//func (c *Client) UpdateConnectArtifact(id string, updateArtifactRequest camv1.CamV1ConnectArtifactUpdate) (camv1.CamV1ConnectArtifact, error) {
//	resp, httpResp, err := c.ConnectArtifactClient.ConnectArtifactsCamV1Api.UpdateCamV1ConnectArtifact(c.connectArtifactApiContext(), id).CamV1ConnectArtifactUpdate(updateArtifactRequest).Execute()
//	return resp, errors.CatchCCloudV2Error(err, httpResp)
//}

func (c *Client) executeListConnectArtifacts(pageToken, cloud, region, env string) (camv1.CamV1ConnectArtifactList, *http.Response, error) {
	req := c.ConnectArtifactClient.ConnectArtifactsCamV1Api.ListCamV1ConnectArtifacts(c.connectArtifactApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	if cloud != "" {
		req = req.SpecCloud(cloud)
	}
	if region != "" {
		req = req.SpecRegion(region)
	}
	if env != "" {
		req = req.Environment(env)
	}
	return req.Execute()
}
