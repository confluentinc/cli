package ccloudv2

import (
	"context"
	"net/http"

	billingv1 "github.com/confluentinc/ccloud-sdk-go-v2/billing/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newBillingClient(url, userAgent string, unsafeTrace bool) *billingv1.APIClient {
	cfg := billingv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
	cfg.Servers = billingv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return billingv1.NewAPIClient(cfg)
}

func (c *Client) billingApiContext() context.Context {
	return context.WithValue(context.Background(), billingv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) ListBillingCosts(startDate, endDate string) ([]billingv1.BillingV1Cost, error) {
	var list []billingv1.BillingV1Cost

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListCosts(startDate, endDate, pageToken)
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

func (c *Client) executeListCosts(startDate, endDate, pageToken string) (billingv1.BillingV1CostList, *http.Response, error) {
	req := c.BillingClient.CostsBillingV1Api.ListBillingV1Costs(c.billingApiContext()).PageSize(10000).StartDate(startDate).EndDate(endDate)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
