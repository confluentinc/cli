package flink

import (
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type catalogOut struct {
	CreationTime string   `human:"Creation Time" serialized:"creation_time"`
	Name         string   `human:"Name" serialized:"name"`
	Databases    []string `human:"Databases" serialized:"databases"`
}

func (c *command) newCatalogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "catalog",
		Short:       "Manage Flink catalogs in Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newCatalogCreateCommand())
	cmd.AddCommand(c.newCatalogDeleteCommand())
	cmd.AddCommand(c.newCatalogDescribeCommand())
	cmd.AddCommand(c.newCatalogListCommand())

	return cmd
}

func convertSdkCatalogToLocalCatalog(sdkOutputCatalog cmfsdk.KafkaCatalog) LocalKafkaCatalog {
	localClusters := make([]LocalKafkaCatalogSpecKafkaClusters, 0, len(sdkOutputCatalog.Spec.GetKafkaClusters()))
	for _, sdkCluster := range sdkOutputCatalog.Spec.GetKafkaClusters() {
		localClusters = append(localClusters, LocalKafkaCatalogSpecKafkaClusters{
			DatabaseName:       sdkCluster.DatabaseName,
			ConnectionConfig:   sdkCluster.ConnectionConfig,
			ConnectionSecretId: sdkCluster.ConnectionSecretId,
		})
	}

	return LocalKafkaCatalog{
		ApiVersion: sdkOutputCatalog.ApiVersion,
		Kind:       sdkOutputCatalog.Kind,
		Metadata: LocalCatalogMetadata{
			Name:              sdkOutputCatalog.Metadata.Name,
			CreationTimestamp: sdkOutputCatalog.Metadata.CreationTimestamp,
			Uid:               sdkOutputCatalog.Metadata.Uid,
			Labels:            sdkOutputCatalog.Metadata.Labels,
			Annotations:       sdkOutputCatalog.Metadata.Annotations,
		},
		Spec: LocalKafkaCatalogSpec{
			SrInstance: LocalKafkaCatalogSpecSrInstance{
				ConnectionConfig:   sdkOutputCatalog.Spec.SrInstance.ConnectionConfig,
				ConnectionSecretId: sdkOutputCatalog.Spec.SrInstance.ConnectionSecretId,
			},
			KafkaClusters: localClusters,
		},
	}
}
