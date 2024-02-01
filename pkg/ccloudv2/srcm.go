package ccloudv2

import (
	"context"
	"net/http"

	srcmv2 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"
	srcmv3 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v3"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newSrcmV2Client(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *srcmv2.APIClient {
	cfg := srcmv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = srcmv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return srcmv2.NewAPIClient(cfg)
}

func newSrcmV3Client(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *srcmv3.APIClient {
	cfg := srcmv3.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = srcmv3.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return srcmv3.NewAPIClient(cfg)
}

func (c *Client) srcmApiContext() context.Context {
	return context.WithValue(context.Background(), srcmv2.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) srcmV3ApiContext() context.Context {
	return context.WithValue(context.Background(), srcmv3.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) GetStreamGovernanceRegionById(regionId string) (srcmv2.SrcmV2Region, error) {
	region, httpResp, err := c.SrcmV2Client.RegionsSrcmV2Api.GetSrcmV2Region(c.srcmApiContext(), regionId).Execute()
	return region, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateSchemaRegistryCluster(srCluster srcmv2.SrcmV2Cluster) (srcmv2.SrcmV2Cluster, error) {
	createdCluster, httpResp, err := c.SrcmV2Client.ClustersSrcmV2Api.CreateSrcmV2Cluster(c.srcmApiContext()).SrcmV2Cluster(srCluster).Execute()
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
	req := c.SrcmV2Client.RegionsSrcmV2Api.ListSrcmV2Regions(c.srcmApiContext())
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
	cluster, httpResp, err := c.SrcmV2Client.ClustersSrcmV2Api.GetSrcmV2Cluster(c.srcmApiContext(), clusterId).Environment(environment).Execute()
	return cluster, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteSchemaRegistryCluster(clusterId, environment string) error {
	httpResp, err := c.SrcmV2Client.ClustersSrcmV2Api.DeleteSrcmV2Cluster(c.srcmApiContext(), clusterId).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpgradeSchemaRegistryCluster(srcmV2ClusterUpdate srcmv2.SrcmV2ClusterUpdate, clusterId string) (srcmv2.SrcmV2Cluster, error) {
	cluster, httpResp, err := c.SrcmV2Client.ClustersSrcmV2Api.UpdateSrcmV2Cluster(c.srcmApiContext(), clusterId).SrcmV2ClusterUpdate(srcmV2ClusterUpdate).Execute()
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

func (c *Client) GetSrcmV3ClustersByEnvironment(environment string) ([]srcmv3.SrcmV3Cluster, error) {
	var list []srcmv3.SrcmV3Cluster

	done := false
	pageToken := ""
	for !done {
		req := c.SrcmV3Client.ClustersSrcmV3Api.ListSrcmV3Clusters(c.srcmV3ApiContext()).Environment(environment)
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}
		page, httpResp, err := c.SrcmV3Client.ClustersSrcmV3Api.ListSrcmV3ClustersExecute(req)
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
	req := c.SrcmV2Client.ClustersSrcmV2Api.ListSrcmV2Clusters(c.srcmApiContext()).Environment(environment)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.SrcmV2Client.ClustersSrcmV2Api.ListSrcmV2ClustersExecute(req)
}
