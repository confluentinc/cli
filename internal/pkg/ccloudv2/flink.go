package ccloudv2

import (
	"context"
	"net/http"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newFlinkClient(url, userAgent string, unsafeTrace bool) *flinkv2.APIClient {
	cfg := flinkv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = flinkv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return flinkv2.NewAPIClient(cfg)
}

func (c *Client) flinkApiContext() context.Context {
	return context.WithValue(context.Background(), flinkv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateFlinkResourcePool(resourcePool flinkv2.FrpmV2ResourcePool) (flinkv2.FrpmV2ResourcePool, error) {
	req := c.FlinkClient.ResourcePoolsFrpmV2Api.CreateFrpmV2ResourcePool(c.flinkApiContext()).FrpmV2ResourcePool(resourcePool)
	res, r, err := c.FlinkClient.ResourcePoolsFrpmV2Api.CreateFrpmV2ResourcePoolExecute(req)
	return res, errors.CatchCCloudV2Error(err, r)
}

func (c *Client) DescribeFlinkResourcePool(id string) (flinkv2.FrpmV2ResourcePool, error) {
	req := c.FlinkClient.ResourcePoolsFrpmV2Api.GetFrpmV2ResourcePool(c.cmkApiContext(), id)
	res, r, err := c.FlinkClient.ResourcePoolsFrpmV2Api.GetFrpmV2ResourcePoolExecute(req)
	return res, errors.CatchCCloudV2Error(err, r)
}

func (c *Client) UpdateFlinkResourcePool(id string, update flinkv2.FrpmV2ResourcePoolUpdate) (flinkv2.FrpmV2ResourcePool, error) {
	req := c.FlinkClient.ResourcePoolsFrpmV2Api.UpdateFrpmV2ResourcePool(c.cmkApiContext(), id).FrpmV2ResourcePoolUpdate(update)
	res, r, err := c.FlinkClient.ResourcePoolsFrpmV2Api.UpdateFrpmV2ResourcePoolExecute(req)
	return res, errors.CatchCCloudV2Error(err, r)
}

func (c *Client) DeleteFlinkResourcePool(id string) error {
	req := c.FlinkClient.ResourcePoolsFrpmV2Api.DeleteFrpmV2ResourcePool(c.cmkApiContext(), id)
	r, err := c.FlinkClient.ResourcePoolsFrpmV2Api.DeleteFrpmV2ResourcePoolExecute(req)
	return errors.CatchCCloudV2Error(err, r)
}

func (c *Client) ListFlinkResourcePools() ([]flinkv2.FrpmV2ResourcePool, error) {
	var list []flinkv2.FrpmV2ResourcePool

	done := false
	pageToken := ""
	for !done {
		page, r, err := c.executeListResourcePools(pageToken)
		if err != nil {
			return nil, errors.CatchCCloudV2Error(err, r)
		}
		list = append(list, page.GetData()...)

		pageToken, done, err = extractFlinkNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) executeListResourcePools(pageToken string) (flinkv2.FrpmV2ResourcePoolList, *http.Response, error) {
	req := c.FlinkClient.ResourcePoolsFrpmV2Api.ListFrpmV2ResourcePools(c.cmkApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.FlinkClient.ResourcePoolsFrpmV2Api.ListFrpmV2ResourcePoolsExecute(req)
}

func extractFlinkNextPageToken(nextPageUrlStringNullable flinkv2.NullableString) (string, bool, error) {
	if !nextPageUrlStringNullable.IsSet() {
		return "", true, nil
	}
	nextPageUrlString := *nextPageUrlStringNullable.Get()
	pageToken, err := extractPageToken(nextPageUrlString)
	return pageToken, false, err
}
