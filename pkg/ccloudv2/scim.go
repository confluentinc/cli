package ccloudv2

import (
	"net/http"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// SCIM token API calls

func (c *Client) CreateScimToken(token orgv2.InlineObject) (orgv2.OrgV2ScimToken, error) {
	resp, httpResp, err := c.OrgClient.ScimTokensOrgV2Api.CreateOrgV2ScimToken(c.orgApiContext()).InlineObject(token).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListScimTokens() ([]orgv2.OrgV2ScimToken, error) {
	var list []orgv2.OrgV2ScimToken

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListScimTokens(pageToken)
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

func (c *Client) executeListScimTokens(pageToken string) (orgv2.OrgV2ScimTokenList, *http.Response, error) {
	req := c.OrgClient.ScimTokensOrgV2Api.ListOrgV2ScimTokens(c.orgApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) DeleteScimToken(id string) error {
	httpResp, err := c.OrgClient.ScimTokensOrgV2Api.DeleteOrgV2ScimToken(c.orgApiContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}
