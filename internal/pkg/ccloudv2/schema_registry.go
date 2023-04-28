package ccloudv2

import (
	"context"
	"net/http"

	srcmv2 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newSchemaRegistryClient(url, userAgent string, unsafeTrace bool) *srcmv2.APIClient {
	cfg := srcmv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = srcmv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return srcmv2.NewAPIClient(cfg)
}

func (c *Client) SchemaRegistryApiContext() context.Context {
	return context.WithValue(context.Background(), srcmv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) GetStreamGovernanceRegionById(regionId string) (srcmv2.SrcmV2Region, error) {
	req := c.SchemaRegistryClient.RegionsSrcmV2Api.GetSrcmV2Region(c.SchemaRegistryApiContext(), regionId)
	region, httpResp, err := c.SchemaRegistryClient.RegionsSrcmV2Api.GetSrcmV2RegionExecute(req)
	return region, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateSchemaRegistryCluster(srCluster srcmv2.SrcmV2Cluster) (srcmv2.SrcmV2Cluster, error) {
	req := c.SchemaRegistryClient.ClustersSrcmV2Api.CreateSrcmV2Cluster(c.SchemaRegistryApiContext()).SrcmV2Cluster(srCluster)
	createdCluster, httpResp, err := c.SchemaRegistryClient.ClustersSrcmV2Api.CreateSrcmV2ClusterExecute(req)
	return createdCluster, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListSchemaRegistryRegions(cloud, packageType string) ([]srcmv2.SrcmV2Region, error) {
	regionListRequest := c.SchemaRegistryClient.RegionsSrcmV2Api.ListSrcmV2Regions(c.SchemaRegistryApiContext())

	if cloud != "" {
		regionListRequest = regionListRequest.SpecCloud(cloud)
	}

	if packageType != "" {
		regionListRequest = regionListRequest.SpecPackages([]string{packageType})
	}

	var regionList []srcmv2.SrcmV2Region
	done := false
	pageToken := ""
	for !done {
		regionListRequest = regionListRequest.PageToken(pageToken)
		regionPage, httpResp, err := c.SchemaRegistryClient.RegionsSrcmV2Api.ListSrcmV2RegionsExecute(regionListRequest)
		if err != nil {
			return nil, errors.CatchCCloudV2Error(err, httpResp)
		}
		regionList = append(regionList, regionPage.GetData()...)

		pageToken, done, err = extractNextPageToken(regionPage.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return regionList, nil
}

func (c *Client) GetSchemaRegistryClusterById(clusterId, environment string) (srcmv2.SrcmV2Cluster, error) {
	cluster, httpResp, err := c.SchemaRegistryClient.ClustersSrcmV2Api.GetSrcmV2Cluster(c.SchemaRegistryApiContext(), clusterId).Environment(environment).Execute()
	return cluster, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteSchemaRegistryCluster(clusterId, environment string) error {
	httpResp, err := c.SchemaRegistryClient.ClustersSrcmV2Api.DeleteSrcmV2Cluster(c.SchemaRegistryApiContext(), clusterId).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpgradeSchemaRegistryCluster(srcmV2ClusterUpdate srcmv2.SrcmV2ClusterUpdate, clusterId string) (srcmv2.SrcmV2Cluster, error) {
	cluster, httpResp, err := c.SchemaRegistryClient.ClustersSrcmV2Api.UpdateSrcmV2Cluster(c.SchemaRegistryApiContext(), clusterId).SrcmV2ClusterUpdate(srcmV2ClusterUpdate).Execute()
	return cluster, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetSchemaRegistryClustersByEnvironment(environment string) ([]srcmv2.SrcmV2Cluster, error) {
	var list []srcmv2.SrcmV2Cluster

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListSchemaRegistryClusters(environment, pageToken)
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

func (c *Client) executeListSchemaRegistryClusters(environment, pageToken string) (srcmv2.SrcmV2ClusterList, *http.Response, error) {
	req := c.SchemaRegistryClient.ClustersSrcmV2Api.ListSrcmV2Clusters(c.SchemaRegistryApiContext()).Environment(environment)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.SchemaRegistryClient.ClustersSrcmV2Api.ListSrcmV2ClustersExecute(req)
}
