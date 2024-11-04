package ccloudv2

import (
	"context"
	"net/http"

	iamipfilteringv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam-ip-filtering/v2"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newIamIpFiltering(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *iamipfilteringv2.APIClient {
	cfg := iamipfilteringv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = iamipfilteringv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return iamipfilteringv2.NewAPIClient(cfg)
}

func (c *Client) IamIpFilteringContext() context.Context {
	return context.WithValue(context.Background(), iamipfilteringv2.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) CreateIamIpFilter(ipFilter iamipfilteringv2.IamV2IpFilter) (iamipfilteringv2.IamV2IpFilter, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPFiltersIamV2Api.CreateIamV2IpFilter(c.IamIpFilteringContext()).IamV2IpFilter(ipFilter).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIamIpFilter(id string) error {
	httpResp, err := c.IamIpFilteringClient.IPFiltersIamV2Api.DeleteIamV2IpFilter(c.IamIpFilteringContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamIpFilter(id string) (iamipfilteringv2.IamV2IpFilter, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPFiltersIamV2Api.GetIamV2IpFilter(c.IamIpFilteringContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIamIpFilters(resourceScope string, includeParentScope string) ([]iamipfilteringv2.IamV2IpFilter, error) {
	var list []iamipfilteringv2.IamV2IpFilter
	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListIpFilters(pageToken, resourceScope, includeParentScope)
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

func (c *Client) executeListIpFilters(pageToken string, resourceScope string, includeParentScope string) (iamipfilteringv2.IamV2IpFilterList, *http.Response, error) {
	req := c.IamIpFilteringClient.IPFiltersIamV2Api.ListIamV2IpFilters(c.IamIpFilteringContext()).PageSize(ccloudV2ListPageSize).ResourceScope(resourceScope).IncludeParentScope(includeParentScope)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) UpdateIamIpFilter(filter iamipfilteringv2.IamV2IpFilter, id string) (iamipfilteringv2.IamV2IpFilter, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPFiltersIamV2Api.UpdateIamV2IpFilter(c.IamIpFilteringContext(), id).IamV2IpFilter(filter).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

// iam ip group api calls

func (c *Client) CreateIamIpGroup(ipGroup iamipfilteringv2.IamV2IpGroup) (iamipfilteringv2.IamV2IpGroup, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPGroupsIamV2Api.CreateIamV2IpGroup(c.IamIpFilteringContext()).IamV2IpGroup(ipGroup).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIamIpGroup(id string) error {
	httpResp, err := c.IamIpFilteringClient.IPGroupsIamV2Api.DeleteIamV2IpGroup(c.IamIpFilteringContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamIpGroup(id string) (iamipfilteringv2.IamV2IpGroup, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPGroupsIamV2Api.GetIamV2IpGroup(c.IamIpFilteringContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIamIpGroups() ([]iamipfilteringv2.IamV2IpGroup, error) {
	var list []iamipfilteringv2.IamV2IpGroup

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

func (c *Client) executeListIpGroups(pageToken string) (iamipfilteringv2.IamV2IpGroupList, *http.Response, error) {
	req := c.IamIpFilteringClient.IPGroupsIamV2Api.ListIamV2IpGroups(c.IamIpFilteringContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) UpdateIamIpGroup(group iamipfilteringv2.IamV2IpGroup, id string) (iamipfilteringv2.IamV2IpGroup, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPGroupsIamV2Api.UpdateIamV2IpGroup(c.IamIpFilteringContext(), id).IamV2IpGroup(group).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
