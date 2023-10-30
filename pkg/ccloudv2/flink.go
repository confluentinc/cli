package ccloudv2

import (
	"context"
	"net/http"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/resource"
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

func (c *Client) CreateFlinkComputePool(computePool flinkv2.FcpmV2ComputePool) (flinkv2.FcpmV2ComputePool, error) {
	res, httpResp, err := c.FlinkClient.ComputePoolsFcpmV2Api.CreateFcpmV2ComputePool(c.flinkApiContext()).FcpmV2ComputePool(computePool).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteFlinkComputePool(id, environment string) error {
	httpResp, err := c.FlinkClient.ComputePoolsFcpmV2Api.DeleteFcpmV2ComputePool(c.flinkApiContext(), id).Environment(environment).Execute()
	return errors.CatchComputePoolNotFoundError(err, id, httpResp)
}

func (c *Client) DescribeFlinkComputePool(id, environment string) (flinkv2.FcpmV2ComputePool, error) {
	res, httpResp, err := c.FlinkClient.ComputePoolsFcpmV2Api.GetFcpmV2ComputePool(c.flinkApiContext(), id).Environment(environment).Execute()
	return res, errors.CatchComputePoolNotFoundError(err, id, httpResp)
}

func (c *Client) ListFlinkComputePools(environment, specRegion string) ([]flinkv2.FcpmV2ComputePool, error) {
	req := c.FlinkClient.ComputePoolsFcpmV2Api.ListFcpmV2ComputePools(c.flinkApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if specRegion != "" {
		req = req.SpecRegion(specRegion)
	}
	res, httpResp, err := req.Execute()
	return res.GetData(), errors.CatchCCloudV2ResourceNotFoundError(err, resource.Environment, httpResp)
}

func (c *Client) ListFlinkRegions(cloud string) ([]flinkv2.FcpmV2Region, error) {
	req := c.FlinkClient.RegionsFcpmV2Api.ListFcpmV2Regions(c.flinkApiContext()).PageSize(ccloudV2ListPageSize)
	if cloud != "" {
		req = req.Cloud(cloud)
	}
	res, httpResp, err := req.Execute()
	return res.GetData(), errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateFlinkComputePool(id string, update flinkv2.FcpmV2ComputePoolUpdate) (flinkv2.FcpmV2ComputePool, error) {
	res, httpResp, err := c.FlinkClient.ComputePoolsFcpmV2Api.UpdateFcpmV2ComputePool(c.flinkApiContext(), id).FcpmV2ComputePoolUpdate(update).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}
