package ccloudv2

import (
	"context"

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

func (c *Client) CreateFlinkComputePool(computePool flinkv2.FcpmV2ComputePool) (flinkv2.FcpmV2ComputePool, error) {
	req := c.FlinkClient.ComputePoolsFcpmV2Api.CreateFcpmV2ComputePool(c.flinkApiContext()).FcpmV2ComputePool(computePool)
	res, r, err := c.FlinkClient.ComputePoolsFcpmV2Api.CreateFcpmV2ComputePoolExecute(req)
	return res, errors.CatchCCloudV2Error(err, r)
}

func (c *Client) DescribeFlinkComputePool(id, environment string) (flinkv2.FcpmV2ComputePool, error) {
	req := c.FlinkClient.ComputePoolsFcpmV2Api.GetFcpmV2ComputePool(c.flinkApiContext(), id).Environment(environment)
	res, r, err := c.FlinkClient.ComputePoolsFcpmV2Api.GetFcpmV2ComputePoolExecute(req)
	return res, errors.CatchCCloudV2Error(err, r)
}

func (c *Client) ListFlinkComputePools(specRegion, environment string) ([]flinkv2.FcpmV2ComputePool, error) {
	req := c.FlinkClient.ComputePoolsFcpmV2Api.ListFcpmV2ComputePools(c.flinkApiContext()).SpecRegion(specRegion).Environment(environment).PageSize(ccloudV2ListPageSize)
	res, r, err := c.FlinkClient.ComputePoolsFcpmV2Api.ListFcpmV2ComputePoolsExecute(req)
	return res.GetData(), errors.CatchCCloudV2Error(err, r)
}

func (c *Client) ListFlinkRegions(cloud string) ([]flinkv2.FcpmV2Region, error) {
	req := c.FlinkClient.RegionsFcpmV2Api.ListFcpmV2Regions(c.flinkApiContext()).PageSize(ccloudV2ListPageSize)
	if cloud != "" {
		req = req.Cloud(cloud)
	}
	res, r, err := c.FlinkClient.RegionsFcpmV2Api.ListFcpmV2RegionsExecute(req)
	return res.GetData(), errors.CatchCCloudV2Error(err, r)
}
