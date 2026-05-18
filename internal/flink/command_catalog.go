package flink

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
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
	cmd.AddCommand(c.newCatalogUpdateCommand())

	return cmd
}

func printCatalogOutput(cmd *cobra.Command, sdkOutputCatalog cmfsdk.KafkaCatalog) error {
	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		databases := make([]string, 0, len(sdkOutputCatalog.Spec.GetKafkaClusters()))
		for _, kafkaCluster := range sdkOutputCatalog.Spec.GetKafkaClusters() {
			databases = append(databases, kafkaCluster.DatabaseName)
		}
		var creationTime string
		if sdkOutputCatalog.GetMetadata().CreationTimestamp != nil {
			creationTime = *sdkOutputCatalog.GetMetadata().CreationTimestamp
		}
		table.Add(&catalogOut{
			CreationTime: creationTime,
			Name:         sdkOutputCatalog.GetMetadata().Name,
			Databases:    databases,
		})
		return table.Print()
	}

	localCatalog := convertSdkCatalogToLocalCatalog(sdkOutputCatalog)
	return output.SerializedOutput(cmd, localCatalog)
}

func readCatalogResourceFile(resourceFilePath string) (cmfsdk.KafkaCatalog, error) {
	data, err := os.ReadFile(resourceFilePath)
	if err != nil {
		return cmfsdk.KafkaCatalog{}, fmt.Errorf("failed to read file: %w", err)
	}

	var genericData map[string]interface{}
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &genericData)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &genericData)
	default:
		return cmfsdk.KafkaCatalog{}, errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
	}
	if err != nil {
		return cmfsdk.KafkaCatalog{}, fmt.Errorf("failed to parse input file: %w", err)
	}

	jsonBytes, err := json.Marshal(genericData)
	if err != nil {
		return cmfsdk.KafkaCatalog{}, fmt.Errorf("failed to marshal intermediate data: %w", err)
	}

	var sdkCatalog cmfsdk.KafkaCatalog
	if err = json.Unmarshal(jsonBytes, &sdkCatalog); err != nil {
		return cmfsdk.KafkaCatalog{}, fmt.Errorf("failed to bind data to KafkaCatalog model: %w", err)
	}

	return sdkCatalog, nil
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
