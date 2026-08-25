package ccloudv2

import (
	"context"
	"net/http"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newFlinkClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *flinkv2.APIClient {
	cfg := flinkv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = flinkv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return flinkv2.NewAPIClient(cfg)
}

func (c *Client) flinkApiContext() context.Context {
	return context.WithValue(context.Background(), flinkv2.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) DescribeFlinkComputePool(id, environment string) (flinkv2.FcpmV2ComputePool, error) {
	res, httpResp, err := c.FlinkClient.ComputePoolsFcpmV2Api.GetFcpmV2ComputePool(c.flinkApiContext(), id).Environment(environment).Execute()
	return res, errors.CatchComputePoolNotFoundError(err, id, httpResp)
}

// ===== Flink compute pools API calls =====

func (c *Client) CreateFlinkComputePool(req flinkv2.FcpmV2ComputePool) (flinkv2.FcpmV2ComputePool, *http.Response, error) {
	createReq := c.FlinkClient.ComputePoolsFcpmV2Api.
		CreateFcpmV2ComputePool(c.flinkApiContext()).
		FcpmV2ComputePool(req)
	return createReq.Execute()
}

func (c *Client) GetFlinkComputePool(id string, environment string) (flinkv2.FcpmV2ComputePool, *http.Response, error) {
	getReq := c.FlinkClient.ComputePoolsFcpmV2Api.
		GetFcpmV2ComputePool(c.flinkApiContext(), id)
	getReq = getReq.Environment(environment)
	return getReq.Execute()
}

func (c *Client) UpdateFlinkComputePool(id string, update flinkv2.FcpmV2ComputePoolUpdate) (flinkv2.FcpmV2ComputePool, *http.Response, error) {
	updateReq := c.FlinkClient.ComputePoolsFcpmV2Api.
		UpdateFcpmV2ComputePool(c.flinkApiContext(), id).
		FcpmV2ComputePoolUpdate(update)
	return updateReq.Execute()
}

func (c *Client) DeleteFlinkComputePool(id string, environment string) error {
	deleteReq := c.FlinkClient.ComputePoolsFcpmV2Api.
		DeleteFcpmV2ComputePool(c.flinkApiContext(), id)
	deleteReq = deleteReq.Environment(environment)
	httpResp, err := deleteReq.Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListFlinkComputePools(specRegion string, environment string, specNetwork string) ([]flinkv2.FcpmV2ComputePool, error) {
	var list []flinkv2.FcpmV2ComputePool

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListFlinkComputePools(specRegion, environment, specNetwork, pageToken)
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

func (c *Client) executeListFlinkComputePools(specRegion string, environment string, specNetwork string, pageToken string) (flinkv2.FcpmV2ComputePoolList, *http.Response, error) {
	req := c.FlinkClient.ComputePoolsFcpmV2Api.
		ListFcpmV2ComputePools(c.flinkApiContext()).
		Environment(environment).
		PageSize(ccloudV2ListPageSize)
	if specRegion != "" {
		req = req.SpecRegion(specRegion)
	}
	if specNetwork != "" {
		req = req.SpecNetwork(specNetwork)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

// ===== Flink org compute pool configs API calls =====

func (c *Client) GetFlinkOrgComputePoolConfig() (flinkv2.FcpmV2OrgComputePoolConfig, *http.Response, error) {
	getReq := c.FlinkClient.OrgComputePoolConfigsFcpmV2Api.
		GetFcpmV2OrgComputePoolConfig(c.flinkApiContext())
	return getReq.Execute()
}

func (c *Client) UpdateFlinkOrgComputePoolConfig(update flinkv2.FcpmV2OrgComputePoolConfigUpdate) (flinkv2.FcpmV2OrgComputePoolConfig, *http.Response, error) {
	updateReq := c.FlinkClient.OrgComputePoolConfigsFcpmV2Api.
		UpdateFcpmV2OrgComputePoolConfig(c.flinkApiContext()).
		FcpmV2OrgComputePoolConfigUpdate(update)
	return updateReq.Execute()
}

// ===== Flink regions API calls =====

func (c *Client) ListFlinkRegions(cloud string, regionName string) ([]flinkv2.FcpmV2Region, error) {
	var list []flinkv2.FcpmV2Region

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListFlinkRegions(cloud, regionName, pageToken)
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

func (c *Client) executeListFlinkRegions(cloud string, regionName string, pageToken string) (flinkv2.FcpmV2RegionList, *http.Response, error) {
	req := c.FlinkClient.RegionsFcpmV2Api.
		ListFcpmV2Regions(c.flinkApiContext()).
		PageSize(ccloudV2ListPageSize)
	if cloud != "" {
		req = req.Cloud(cloud)
	}
	if regionName != "" {
		req = req.RegionName(regionName)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
