package ccloudv2

import (
	"context"
	"net/http"

	ssov2 "github.com/confluentinc/ccloud-sdk-go-v2/sso/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newSsoClient(url, userAgent string, unsafeTrace bool) *ssov2.APIClient {
	cfg := ssov2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
	cfg.Servers = ssov2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return ssov2.NewAPIClient(cfg)
}

func (c *Client) groupMappingApiContext() context.Context {
	return context.WithValue(context.Background(), ssov2.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateGroupMapping(groupMapping ssov2.IamV2SsoGroupMapping) (ssov2.IamV2SsoGroupMapping, error) {
	resp, httpResp, err := c.SsoClient.GroupMappingsIamV2SsoApi.CreateIamV2SsoGroupMapping(c.groupMappingApiContext()).IamV2SsoGroupMapping(groupMapping).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteGroupMapping(id string) error {
	httpResp, err := c.SsoClient.GroupMappingsIamV2SsoApi.DeleteIamV2SsoGroupMapping(c.groupMappingApiContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetGroupMapping(id string) (ssov2.IamV2SsoGroupMapping, error) {
	resp, httpResp, err := c.SsoClient.GroupMappingsIamV2SsoApi.GetIamV2SsoGroupMapping(c.groupMappingApiContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateGroupMapping(update ssov2.IamV2SsoGroupMapping) (ssov2.IamV2SsoGroupMapping, error) {
	resp, httpResp, err := c.SsoClient.GroupMappingsIamV2SsoApi.UpdateIamV2SsoGroupMapping(c.groupMappingApiContext(), *update.Id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListGroupMappings() ([]ssov2.IamV2SsoGroupMapping, error) {
	var list []ssov2.IamV2SsoGroupMapping

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListGroupMappings(pageToken)
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

func (c *Client) executeListGroupMappings(pageToken string) (ssov2.IamV2SsoGroupMappingList, *http.Response, error) {
	req := c.SsoClient.GroupMappingsIamV2SsoApi.ListIamV2SsoGroupMappings(c.groupMappingApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
