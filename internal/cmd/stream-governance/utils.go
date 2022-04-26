package streamgovernance

import (
	sgsdk "github.com/confluentinc/ccloud-sdk-go-v2-internal/stream-governance/v1"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
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

func PrintStreamGovernanceClusterOutput(cmd *cobra.Command, newCluster sgsdk.StreamGovernanceV1Cluster) {
	spec := newCluster.GetSpec()
	environment := spec.GetEnvironment()
	status := newCluster.GetStatus()

	clusterResponse := &v1.StreamGovernanceV1Cluster{
		Name:                   spec.GetDisplayName(),
		SchemaRegistryEndpoint: spec.GetHttpEndpoint(),
		Environment:            environment.GetId(),
		Package:                spec.GetPackage(),
		Cloud:                  spec.GetCloud(),
		Region:                 spec.GetRegion(),
		Status:                 status.GetPhase(),
	}

	_ = output.DescribeObject(cmd, clusterResponse, clusterResponseLabels, clusterResponseHumanNames, clusterResponseStructuredRenames)
}
