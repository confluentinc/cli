package streamgovernance

import (
	"context"
	sgsdk "github.com/confluentinc/ccloud-sdk-go-v2/stream-governance/v2"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
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

func PrintStreamGovernanceClusterOutput(cmd *cobra.Command, newCluster sgsdk.StreamGovernanceV2Cluster) {
	spec := newCluster.GetSpec()
	environment := spec.GetEnvironment()
	region := spec.GetRegion()
	status := newCluster.GetStatus()

	//TODO: get correct cloud and region_name from RegionObj
	clusterResponse := &streamGovernanceV2Cluster{
		Name:                   spec.GetDisplayName(),
		SchemaRegistryEndpoint: spec.GetHttpEndpoint(),
		Environment:            environment.GetId(),
		Package:                spec.GetPackage(),
		Cloud:                  region.GetId(),
		Region:                 region.GetId(),
		Status:                 status.GetPhase(),
	}

	_ = output.DescribeObject(cmd, clusterResponse, clusterResponseLabels, clusterResponseHumanNames, clusterResponseStructuredRenames)
}

func (c *streamGovernanceCommand) getClusterIdFromEnvironment(context context.Context) (string, error) {
	ctxClient := dynamicconfig.NewContextClient(c.Context)
	clusterInfo, err := ctxClient.FetchSchemaRegistryByAccountId(context, c.EnvironmentId())
	if err != nil {
		return "", err
	}

	clusterId := clusterInfo.GetId()
	return clusterId, nil
}
