package ccloudv2

import (
	"context"
	"net/http"

	cclv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccl/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newCclClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *cclv1.APIClient {
	cfg := cclv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = cclv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return cclv1.NewAPIClient(cfg)
}

func (c *Client) connectCustomCodeLoggingApiContext() context.Context {
	return context.WithValue(context.Background(), cclv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) CreateCustomCodeLogging(createCustomCodeLoggingRequest cclv1.CclV1CustomCodeLogging) (cclv1.CclV1CustomCodeLogging, error) {
	resp, httpResp, err := c.Cclv1Client.CustomCodeLoggingsCclV1Api.CreateCclV1CustomCodeLogging(c.connectCustomCodeLoggingApiContext()).CclV1CustomCodeLogging(createCustomCodeLoggingRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListCustomCodeLoggings(env string) ([]cclv1.CclV1CustomCodeLogging, error) {
	var list []cclv1.CclV1CustomCodeLogging

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListCustomCodeLoggings(pageToken, env)
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

func (c *Client) DescribeCustomCodeLogging(id string, env string) (cclv1.CclV1CustomCodeLogging, error) {
	resp, httpResp, err := c.Cclv1Client.CustomCodeLoggingsCclV1Api.GetCclV1CustomCodeLogging(c.connectCustomCodeLoggingApiContext(), id).Environment(env).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteCustomCodeLogging(id string, env string) error {
	httpResp, err := c.Cclv1Client.CustomCodeLoggingsCclV1Api.DeleteCclV1CustomCodeLogging(c.connectCustomCodeLoggingApiContext(), id).Environment(env).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateCustomCodeLogging(id string, env string, updateCustomCodeLoggingRequest cclv1.CclV1CustomCodeLoggingUpdate) (cclv1.CclV1CustomCodeLogging, error) {
	resp, httpResp, err := c.Cclv1Client.CustomCodeLoggingsCclV1Api.UpdateCclV1CustomCodeLogging(c.connectCustomCodeLoggingApiContext(), id).CclV1CustomCodeLoggingUpdate(updateCustomCodeLoggingRequest).Environment(env).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) executeListCustomCodeLoggings(pageToken string, env string) (cclv1.CclV1CustomCodeLoggingList, *http.Response, error) {
	req := c.Cclv1Client.CustomCodeLoggingsCclV1Api.ListCclV1CustomCodeLoggings(c.connectCustomCodeLoggingApiContext()).Environment(env).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
