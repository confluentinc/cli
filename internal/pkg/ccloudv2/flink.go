package ccloudv2

import (
	"context"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type ListFlinkIAMBindingsQueryParams struct {
	Region         string
	Cloud          string
	IdentityPoolId string
}

func (p *ListFlinkIAMBindingsQueryParams) GetRegion() string {
	if p == nil {
		return ""
	}
	return p.Region
}

func (p *ListFlinkIAMBindingsQueryParams) GetCloud() string {
	if p == nil {
		return ""
	}
	return p.Cloud
}

func (p *ListFlinkIAMBindingsQueryParams) GetIdentityPoolId() string {
	if p == nil {
		return ""
	}
	return p.IdentityPoolId
}

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
	res, httpResp, err := c.FlinkClient.ComputePoolsFcpmV2Api.CreateFcpmV2ComputePool(c.flinkApiContext()).FcpmV2ComputePool(computePool).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteFlinkComputePool(id, environment string) error {
	httpResp, err := c.FlinkClient.ComputePoolsFcpmV2Api.DeleteFcpmV2ComputePool(c.flinkApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeFlinkComputePool(id, environment string) (flinkv2.FcpmV2ComputePool, error) {
	res, httpResp, err := c.FlinkClient.ComputePoolsFcpmV2Api.GetFcpmV2ComputePool(c.flinkApiContext(), id).Environment(environment).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListFlinkComputePools(environment, specRegion string) ([]flinkv2.FcpmV2ComputePool, error) {
	req := c.FlinkClient.ComputePoolsFcpmV2Api.ListFcpmV2ComputePools(c.flinkApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if specRegion != "" {
		req = req.SpecRegion(specRegion)
	}
	res, httpResp, err := req.Execute()
	return res.GetData(), errors.CatchCCloudV2Error(err, httpResp)
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

func (c *Client) CreateFlinkIAMBinding(region, cloud, environmentId, identityPoolId string) (flinkv2.FcpmV2IamBinding, error) {
	iamBinding := flinkv2.FcpmV2IamBinding{
		Region:       flinkv2.PtrString(region),
		Cloud:        flinkv2.PtrString(cloud),
		Environment:  flinkv2.NewGlobalObjectReference(environmentId, "", ""),
		IdentityPool: flinkv2.NewGlobalObjectReference(identityPoolId, "", ""),
	}
	res, httpResp, err := c.FlinkClient.IamBindingsFcpmV2Api.CreateFcpmV2IamBinding(c.flinkApiContext()).FcpmV2IamBinding(iamBinding).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteFlinkIAMBinding(id, environmentId string) error {
	httpResp, err := c.FlinkClient.IamBindingsFcpmV2Api.DeleteFcpmV2IamBinding(c.flinkApiContext(), id).Environment(environmentId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListFlinkIAMBindings(environmentId string, params *ListFlinkIAMBindingsQueryParams) ([]flinkv2.FcpmV2IamBinding, error) {
	req := c.FlinkClient.IamBindingsFcpmV2Api.ListFcpmV2IamBindings(c.flinkApiContext()).Environment(environmentId).PageSize(ccloudV2ListPageSize)
	if params.GetRegion() != "" {
		req = req.Region(params.Region)
	}
	if params.GetCloud() != "" {
		req = req.Cloud(params.Cloud)
	}
	if params.GetIdentityPoolId() != "" {
		req = req.IdentityPool(params.IdentityPoolId)
	}
	res, httpResp, err := req.Execute()
	return res.GetData(), errors.CatchCCloudV2Error(err, httpResp)
}
