package streamgovernance

import (
	"context"
	sgsdk "github.com/confluentinc/ccloud-sdk-go-v2/stream-governance/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

var (
	clusterResponseLabels = []string{"Name", "SchemaRegistryEndpoint", "Package", "Environment", "Cloud",
		"Region", "Status"}
	clusterResponseHumanNames = map[string]string{"Name": "Display Name", "SchemaRegistryEndpoint": "Endpoint URL",
		"Environment": "Environment", "Package": "Package", "Cloud": "Cloud", "Region": "Region", "Status": "Status"}
	clusterResponseStructuredRenames = map[string]string{"Name": "display_name",
		"SchemaRegistryEndpoint": "endpoint_url", "Environment": "environment", "Package": "package", "Cloud": "cloud",
		"Region": "region", "Status": "status"}
)

func (c *streamGovernanceCommand) getStreamGovernanceV2Region(cloud, region, packageType string) (*sgsdk.StreamGovernanceV2Region, error) {
	ctx := context.Background()

	packageSpec := sgsdk.NewMultipleSearchFilter()
	packageSpec.Items = append(packageSpec.Items, packageType)
	regionList, _, err := c.V2Client.StreamGovernanceClient.RegionsStreamGovernanceV2Api.ListStreamGovernanceV2Regions(ctx).
		SpecCloud(cloud).SpecRegionName(region).SpecPackages(*packageSpec).Execute()

	if err != nil {
		return nil, err
	}

	regionArr := regionList.GetData()
	if len(regionArr) == 0 {
		return nil, errors.NewErrorWithSuggestions(errors.SGInvalidRegionErrorMsg,
			errors.SGInvalidRegionSuggestions)
	}

	return &regionArr[0], nil
}

func (c *streamGovernanceCommand) getStreamGovernanceV2ClusterIdForEnvironment(context context.Context) (string, error) {
	cluster, err := c.getStreamGovernanceV2ClusterForEnvironment(context)
	if err != nil {
		return "", err
	}

	return cluster.GetId(), nil
}

func (c *streamGovernanceCommand) getStreamGovernanceV2ClusterForEnvironment(context context.Context) (*sgsdk.StreamGovernanceV2Cluster, error) {
	clusterList, _, err := c.V2Client.StreamGovernanceClient.ClustersStreamGovernanceV2Api.
		ListStreamGovernanceV2Clusters(context).Environment(c.EnvironmentId()).Execute()

	if err != nil {
		return nil, err
	}

	clusterArr := clusterList.GetData()
	if len(clusterArr) == 0 {
		return nil, errors.NewStreamGovernanceNotEnabledError()
	}

	return &clusterArr[0], nil
}

func (c *streamGovernanceCommand) getStreamGovernanceV2RegionFromId(regionId string) (*sgsdk.StreamGovernanceV2Region, error) {
	ctx := context.Background()

	regionObject, _, err := c.V2Client.StreamGovernanceClient.
		RegionsStreamGovernanceV2Api.GetStreamGovernanceV2Region(ctx, regionId).Execute()
	if err != nil {
		return nil, err
	}

	return &regionObject, nil
}

func PrintStreamGovernanceClusterOutput(cmd *cobra.Command, newCluster sgsdk.StreamGovernanceV2Cluster, region sgsdk.StreamGovernanceV2Region) {
	spec := newCluster.GetSpec()
	environment := spec.GetEnvironment()
	regionSpec := region.GetSpec()
	status := newCluster.GetStatus()

	clusterResponse := &streamGovernanceV2Cluster{
		Name:                   spec.GetDisplayName(),
		SchemaRegistryEndpoint: spec.GetHttpEndpoint(),
		Environment:            environment.GetId(),
		Package:                spec.GetPackage(),
		Cloud:                  regionSpec.GetCloud(),
		Region:                 regionSpec.GetDisplayName(),
		Status:                 status.GetPhase(),
	}

	_ = output.DescribeObject(cmd, clusterResponse, clusterResponseLabels, clusterResponseHumanNames, clusterResponseStructuredRenames)
}
