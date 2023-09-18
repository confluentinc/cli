package ccloudv2

import (
	"context"

	networkingprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-privatelink/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newNetworkingPrivateLinkClient(url, userAgent string, unsafeTrace bool) *networkingprivatelinkv1.APIClient {
	cfg := networkingprivatelinkv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
	cfg.Servers = networkingprivatelinkv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return networkingprivatelinkv1.NewAPIClient(cfg)
}

func (c *Client) networkingPrivateLinkApiContext() context.Context {
	return context.WithValue(context.Background(), networkingprivatelinkv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) ListPrivateLinkAttachments(environment string) ([]networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment, error) {
	var list []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListPrivateLinkAttachments(environment, pageToken)
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

func (c *Client) executeListPrivateLinkAttachments(environment, pageToken string) (networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentList, error) {
	req := c.NetworkingPrivateLinkClient.PrivateLinkAttachmentsNetworkingV1Api.ListNetworkingV1PrivateLinkAttachments(c.networkingPrivateLinkApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetPrivateLinkAttachment(environment, id string) (networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment, error) {
	resp, httpResp, err := c.NetworkingPrivateLinkClient.PrivateLinkAttachmentsNetworkingV1Api.GetNetworkingV1PrivateLinkAttachment(c.networkingPrivateLinkApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
