package streamgovernance

import (
	sgsdk "github.com/confluentinc/ccloud-sdk-go-v2-internal/stream-governance/v1"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

var (
	clusterResponseLabels = []string{"APIVersion", "Id", "Kind", "Resource Name", "SchemaRegistryEndpoint",
		"Package", "Environment", "Cloud", "Region", "Status"}
	clusterResponseHumanNames = map[string]string{"APIVersion": "API Version", "ID": "Cluster ID", "Kind": "Kind",
		"SchemaRegistryEndpoint": "Endpoint URL", "Environment": "Environment", "Package": "Package",
		"Cloud": "Cloud", "Region": "Region", "Status": "Status"}
	clusterResponseStructuredRenames = map[string]string{"APIVersion": "api_Version", "ID": "id", "Kind": "kind",
		"SchemaRegistryEndpoint": "endpoint_url", "Environment": "environment", "Package": "package",
		"Cloud": "cloud", "Region": "region", "Status": "status"}
)

func PrintStreamGovernanceClusterOutput(cmd *cobra.Command, newCluster sgsdk.StreamGovernanceV1Cluster) {
	metadata := newCluster.GetMetadata()
	spec := newCluster.GetSpec()
	environment := spec.GetEnvironment()
	status := newCluster.GetStatus()

	clusterResponse := &v1.StreamGovernanceV1Cluster{
		APIVersion:             newCluster.GetApiVersion(),
		Id:                     newCluster.GetId(),
		Kind:                   newCluster.GetKind(),
		ResourceName:           metadata.GetResourceName(),
		SchemaRegistryEndpoint: spec.GetHttpEndpoint(),
		Environment:            environment.GetId(),
		Package:                spec.GetPackage(),
		Cloud:                  spec.GetCloud(),
		Region:                 spec.GetRegion(),
		Status:                 status.GetPhase(),
	}

	_ = output.DescribeObject(cmd, clusterResponse, clusterResponseLabels, clusterResponseHumanNames, clusterResponseStructuredRenames)
}
