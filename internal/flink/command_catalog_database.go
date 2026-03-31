package flink

import (
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type databaseOut struct {
	CreationTime string `human:"Creation Time" serialized:"creation_time"`
	Name         string `human:"Name" serialized:"name"`
	Catalog      string `human:"Catalog" serialized:"catalog"`
}

func (c *command) newCatalogDatabaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "database",
		Short:       "Manage Flink databases in Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newCatalogDatabaseCreateCommand())
	cmd.AddCommand(c.newCatalogDatabaseDescribeCommand())
	cmd.AddCommand(c.newCatalogDatabaseListCommand())
	cmd.AddCommand(c.newCatalogDatabaseUpdateCommand())

	return cmd
}

func convertSdkDatabaseToLocalDatabase(sdkDatabase cmfsdk.KafkaDatabase) LocalKafkaDatabase {
	return LocalKafkaDatabase{
		ApiVersion: sdkDatabase.ApiVersion,
		Kind:       sdkDatabase.Kind,
		Metadata: LocalDatabaseMetadata{
			Name:              sdkDatabase.Metadata.Name,
			CreationTimestamp: sdkDatabase.Metadata.CreationTimestamp,
			UpdateTimestamp:   sdkDatabase.Metadata.UpdateTimestamp,
			Uid:               sdkDatabase.Metadata.Uid,
			Labels:            sdkDatabase.Metadata.Labels,
			Annotations:       sdkDatabase.Metadata.Annotations,
		},
		Spec: LocalKafkaDatabaseSpec{
			KafkaCluster: LocalKafkaDatabaseSpecKafkaCluster{
				ConnectionConfig:   sdkDatabase.Spec.KafkaCluster.ConnectionConfig,
				ConnectionSecretId: sdkDatabase.Spec.KafkaCluster.ConnectionSecretId,
			},
			AlterEnvironments: sdkDatabase.Spec.AlterEnvironments,
		},
	}
}
