package ccloudv2

import (
	"context"
	"net/http"

	srcmv2 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newSrcmClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *srcmv2.APIClient {
	cfg := srcmv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = srcmv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return srcmv2.NewAPIClient(cfg)
}

func (c *Client) srcmApiContext() context.Context {
	return context.WithValue(context.Background(), srcmv2.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) GetStreamGovernanceRegionById(regionId string) (srcmv2.SrcmV2Region, error) {
	region, httpResp, err := c.SrcmClient.RegionsSrcmV2Api.GetSrcmV2Region(c.srcmApiContext(), regionId).Execute()
	return region, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateSchemaRegistryCluster(srCluster srcmv2.SrcmV2Cluster) (srcmv2.SrcmV2Cluster, error) {
	createdCluster, httpResp, err := c.SrcmClient.ClustersSrcmV2Api.CreateSrcmV2Cluster(c.srcmApiContext()).SrcmV2Cluster(srCluster).Execute()
	return createdCluster, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListSchemaRegistryRegions(cloud, packageType string) ([]srcmv2.SrcmV2Region, error) {
	var list []srcmv2.SrcmV2Region

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListSchemaRegistryRegions(cloud, packageType, pageToken)
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

func (c *Client) executeListSchemaRegistryRegions(cloud, packageType, pageToken string) (srcmv2.SrcmV2RegionList, *http.Response, error) {
	req := c.SrcmClient.RegionsSrcmV2Api.ListSrcmV2Regions(c.srcmApiContext())
	if cloud != "" {
		req = req.SpecCloud(cloud)
	}
	if packageType != "" {
		req = req.SpecPackages([]string{packageType})
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) GetSchemaRegistryClusterById(clusterId, environment string) (srcmv2.SrcmV2Cluster, error) {
	cluster, httpResp, err := c.SrcmClient.ClustersSrcmV2Api.GetSrcmV2Cluster(c.srcmApiContext(), clusterId).Environment(environment).Execute()
	return cluster, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteSchemaRegistryCluster(clusterId, environment string) error {
	httpResp, err := c.SrcmClient.ClustersSrcmV2Api.DeleteSrcmV2Cluster(c.srcmApiContext(), clusterId).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpgradeSchemaRegistryCluster(srcmV2ClusterUpdate srcmv2.SrcmV2ClusterUpdate, clusterId string) (srcmv2.SrcmV2Cluster, error) {
	cluster, httpResp, err := c.SrcmClient.ClustersSrcmV2Api.UpdateSrcmV2Cluster(c.srcmApiContext(), clusterId).SrcmV2ClusterUpdate(srcmV2ClusterUpdate).Execute()
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
	req := c.SrcmClient.ClustersSrcmV2Api.ListSrcmV2Clusters(c.srcmApiContext()).Environment(environment)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.SrcmClient.ClustersSrcmV2Api.ListSrcmV2ClustersExecute(req)
}
