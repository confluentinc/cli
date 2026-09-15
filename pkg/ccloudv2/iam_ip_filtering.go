package ccloudv2

import (
	"context"
	"net/http"

	iamipfilteringv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam-ip-filtering/v2"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// ===== API group client bootstrap =====

func newIamIpFilteringClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *iamipfilteringv2.APIClient {
	cfg := iamipfilteringv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = iamipfilteringv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return iamipfilteringv2.NewAPIClient(cfg)
}

func (c *Client) iamIpFilteringApiContext() context.Context {
	return context.WithValue(context.Background(), iamipfilteringv2.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) CreateIamIpFilter(ipFilter iamipfilteringv2.IamV2IpFilter) (iamipfilteringv2.IamV2IpFilter, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPFiltersIamV2Api.CreateIamV2IpFilter(c.iamIpFilteringApiContext()).IamV2IpFilter(ipFilter).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIamIpFilter(id string) error {
	httpResp, err := c.IamIpFilteringClient.IPFiltersIamV2Api.DeleteIamV2IpFilter(c.iamIpFilteringApiContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamIpFilter(id string) (iamipfilteringv2.IamV2IpFilter, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPFiltersIamV2Api.GetIamV2IpFilter(c.iamIpFilteringApiContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIamIpFilters(resourceScope string, includeParentScopes string) ([]iamipfilteringv2.IamV2IpFilter, error) {
	var list []iamipfilteringv2.IamV2IpFilter
	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListIpFilters(pageToken, resourceScope, includeParentScopes)
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

func (c *Client) executeListIpFilters(pageToken string, resourceScope string, includeParentScopes string) (iamipfilteringv2.IamV2IpFilterList, *http.Response, error) {
	req := c.IamIpFilteringClient.IPFiltersIamV2Api.ListIamV2IpFilters(c.iamIpFilteringApiContext()).PageSize(ccloudV2ListPageSize).ResourceScope(resourceScope).IncludeParentScopes(includeParentScopes)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) UpdateIamIpFilter(filter iamipfilteringv2.IamV2IpFilter, id string) (iamipfilteringv2.IamV2IpFilter, error) {
	resp, httpResp, err := c.IamIpFilteringClient.IPFiltersIamV2Api.UpdateIamV2IpFilter(c.iamIpFilteringApiContext(), id).IamV2IpFilter(filter).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

// ===== IAM IP filtering IP groups API calls =====

func (c *Client) CreateIamIpGroup(req iamipfilteringv2.IamV2IpGroup) (iamipfilteringv2.IamV2IpGroup, error) {
	createReq := c.IamIpFilteringClient.IPGroupsIamV2Api.
		CreateIamV2IpGroup(c.iamIpFilteringApiContext()).
		IamV2IpGroup(req)
	res, httpResp, err := createReq.Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamIpGroup(id string) (iamipfilteringv2.IamV2IpGroup, error) {
	getReq := c.IamIpFilteringClient.IPGroupsIamV2Api.
		GetIamV2IpGroup(c.iamIpFilteringApiContext(), id)
	res, httpResp, err := getReq.Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateIamIpGroup(id string, update iamipfilteringv2.IamV2IpGroup) (iamipfilteringv2.IamV2IpGroup, error) {
	updateReq := c.IamIpFilteringClient.IPGroupsIamV2Api.
		UpdateIamV2IpGroup(c.iamIpFilteringApiContext(), id).
		IamV2IpGroup(update)
	res, httpResp, err := updateReq.Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIamIpGroup(id string) error {
	deleteReq := c.IamIpFilteringClient.IPGroupsIamV2Api.
		DeleteIamV2IpGroup(c.iamIpFilteringApiContext(), id)
	httpResp, err := deleteReq.Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIamIpGroups() ([]iamipfilteringv2.IamV2IpGroup, error) {
	var list []iamipfilteringv2.IamV2IpGroup

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListIamIpGroups(pageToken)
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

func (c *Client) executeListIamIpGroups(pageToken string) (iamipfilteringv2.IamV2IpGroupList, *http.Response, error) {
	req := c.IamIpFilteringClient.IPGroupsIamV2Api.
		ListIamV2IpGroups(c.iamIpFilteringApiContext()).
		PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
