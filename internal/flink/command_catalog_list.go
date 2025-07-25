package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCatalogListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink catalogs in Confluent Platform.",
		Args:  cobra.NoArgs,
		RunE:  c.catalogList,
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) catalogList(cmd *cobra.Command, _ []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkCatalogs, err := client.ListCatalog(c.createContext())
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, catalog := range sdkCatalogs {
			databases := make([]string, 0, len(catalog.GetSpec().KafkaClusters))
			for _, kafkaCluster := range catalog.GetSpec().KafkaClusters {
				databases = append(databases, kafkaCluster.DatabaseName)
			}
			var creationTime string
			if catalog.GetMetadata().CreationTimestamp != nil {
				creationTime = *catalog.GetMetadata().CreationTimestamp
			} else {
				creationTime = ""
			}
			list.Add(&catalogOut{
				CreationTime: creationTime,
				Name:         catalog.Metadata.Name,
				Databases:    databases,
			})
		}
		return list.Print()
	}

	localCatalogs := make([]LocalKafkaCatalog, 0, len(sdkCatalogs))

	for _, sdkCatalog := range sdkCatalogs {
		localClusters := make([]LocalKafkaCatalogSpecKafkaClusters, 0, len(sdkCatalog.Spec.KafkaClusters))
		for _, sdkCluster := range sdkCatalog.Spec.KafkaClusters {
			localClusters = append(localClusters, LocalKafkaCatalogSpecKafkaClusters{
				DatabaseName:       sdkCluster.DatabaseName,
				ConnectionConfig:   sdkCluster.ConnectionConfig,
				ConnectionSecretId: sdkCluster.ConnectionSecretId,
			})
		}

		localCat := LocalKafkaCatalog{
			ApiVersion: sdkCatalog.ApiVersion,
			Kind:       sdkCatalog.Kind,
			Metadata: LocalCatalogMetadata{
				Name:              sdkCatalog.Metadata.Name,
				CreationTimestamp: sdkCatalog.Metadata.CreationTimestamp,
				Uid:               sdkCatalog.Metadata.Uid,
				Labels:            sdkCatalog.Metadata.Labels,
				Annotations:       sdkCatalog.Metadata.Annotations,
			},
			Spec: LocalKafkaCatalogSpec{
				SrInstance: LocalKafkaCatalogSpecSrInstance{
					ConnectionConfig:   sdkCatalog.Spec.SrInstance.ConnectionConfig,
					ConnectionSecretId: sdkCatalog.Spec.SrInstance.ConnectionSecretId,
				},
				KafkaClusters: localClusters,
			},
		}

		localCatalogs = append(localCatalogs, localCat)
	}

	return output.SerializedOutput(cmd, localCatalogs)
}
