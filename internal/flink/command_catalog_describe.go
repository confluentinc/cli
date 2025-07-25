package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCatalogDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink catalog in Confluent Platform.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.catalogDescribe,
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) catalogDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkOutputCatalog, err := client.DescribeCatalog(c.createContext(), name)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		databases := make([]string, 0, len(sdkOutputCatalog.GetSpec().KafkaClusters))
		for _, kafkaCluster := range sdkOutputCatalog.GetSpec().KafkaClusters {
			databases = append(databases, kafkaCluster.DatabaseName)
		}
		var creationTime string
		if sdkOutputCatalog.GetMetadata().CreationTimestamp != nil {
			creationTime = *sdkOutputCatalog.GetMetadata().CreationTimestamp
		} else {
			creationTime = ""
		}
		table.Add(&catalogOut{
			CreationTime: creationTime,
			Name:         sdkOutputCatalog.GetMetadata().Name,
			Databases:    databases,
		})
		return table.Print()
	}

	localClusters := make([]LocalKafkaCatalogSpecKafkaClusters, 0, len(sdkOutputCatalog.Spec.KafkaClusters))
	for _, sdkCluster := range sdkOutputCatalog.Spec.KafkaClusters {
		localClusters = append(localClusters, LocalKafkaCatalogSpecKafkaClusters{
			DatabaseName:       sdkCluster.DatabaseName,
			ConnectionConfig:   sdkCluster.ConnectionConfig,
			ConnectionSecretId: sdkCluster.ConnectionSecretId,
		})
	}

	localCatalog := LocalKafkaCatalog{
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
	return output.SerializedOutput(cmd, localCatalog)
}
