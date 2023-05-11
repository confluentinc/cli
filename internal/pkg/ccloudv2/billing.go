package ccloudv2

import (
	"context"
	billing "github.com/confluentinc/ccloud-sdk-go-v2/billing/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"net/http"
)

func newBillingClient(url, userAgent string, unsafeTrace bool) *billing.APIClient {
	cfg := billing.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = billing.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return billing.NewAPIClient(cfg)
}

func (c *Client) billingApiContext() context.Context {
	return context.WithValue(context.Background(), billing.ContextAccessToken, c.AuthToken)
}

func (c *Client) ListBillingCosts(startDate, endDate string) ([]billing.BillingV1Cost, error) {
	var list []billing.BillingV1Cost

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

func (c *Client) executeListCosts(startDate, endDate, pageToken string) (billing.BillingV1CostList, *http.Response, error) {
	req := c.BillingClient.CostsBillingV1Api.ListBillingV1Costs(c.billingApiContext()).
		PageSize(10000).
		StartDate(startDate).
		EndDate(endDate)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.BillingClient.CostsBillingV1Api.ListBillingV1CostsExecute(req)
}
