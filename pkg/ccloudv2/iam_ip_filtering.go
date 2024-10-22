package ccloudv2

import (
	"context"
	sdk "github.com/confluentinc/ccloud-sdk-go-v2/iam-ip-filtering/v2"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"net/http"
)

func newIamIpFiltering(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *sdk.APIClient {
	cfg := sdk.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = sdk.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return sdk.NewAPIClient(cfg)
}

func (c *Client) iamIpFilteringContext() context.Context {
	return context.WithValue(context.Background(), sdk.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) CreateIamIpFilter(ipFilter sdk.IamV2IpFilter) (sdk.IamV2IpFilter, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPFiltersIamV2Api.CreateIamV2IpFilter(c.iamIpFilteringContext()).IamV2IpFilter(ipFilter).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIamIpFilter(id string) error {
	httpResp, err := c.IamIpFilteringClient.IPFiltersIamV2Api.DeleteIamV2IpFilter(c.iamIpFilteringContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamIpFilter(id string) (sdk.IamV2IpFilter, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPFiltersIamV2Api.GetIamV2IpFilter(c.iamIpFilteringContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIamIpFilters() ([]sdk.IamV2IpFilter, error) {
	var list []sdk.IamV2IpFilter

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListIpFilters(pageToken)
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

func (c *Client) executeListIpFilters(pageToken string) (sdk.IamV2IpFilterList, *http.Response, error) {
	req := c.IamIpFilteringClient.IPFiltersIamV2Api.ListIamV2IpFilters(c.iamIpFilteringContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) UpdateIamIpFilter(filter sdk.IamV2IpFilter, id string) (sdk.IamV2IpFilter, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPFiltersIamV2Api.UpdateIamV2IpFilter(c.iamIpFilteringContext(), id).IamV2IpFilter(filter).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

// iam ip group api calls

func (c *Client) CreateIamIpGroup(ipGroup sdk.IamV2IpGroup) (sdk.IamV2IpGroup, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPGroupsIamV2Api.CreateIamV2IpGroup(c.iamIpFilteringContext()).IamV2IpGroup(ipGroup).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIamIpGroup(id string) error {
	httpResp, err := c.IamIpFilteringClient.IPGroupsIamV2Api.DeleteIamV2IpGroup(c.iamIpFilteringContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamIpGroup(id string) (sdk.IamV2IpGroup, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPGroupsIamV2Api.GetIamV2IpGroup(c.iamIpFilteringContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIamIpGroups() ([]sdk.IamV2IpGroup, error) {
	var list []sdk.IamV2IpGroup

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListIpGroups(pageToken)
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

func (c *Client) executeListIpGroups(pageToken string) (sdk.IamV2IpGroupList, *http.Response, error) {
	req := c.IamIpFilteringClient.IPGroupsIamV2Api.ListIamV2IpGroups(c.iamIpFilteringContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) UpdateIamIpGroup(group sdk.IamV2IpGroup, id string) (sdk.IamV2IpGroup, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPGroupsIamV2Api.UpdateIamV2IpGroup(c.iamIpFilteringContext(), id).IamV2IpGroup(group).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
