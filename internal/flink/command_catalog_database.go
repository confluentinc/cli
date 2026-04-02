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
	cmd.AddCommand(c.newCatalogDatabaseDeleteCommand())
	cmd.AddCommand(c.newCatalogDatabaseDescribeCommand())
	cmd.AddCommand(c.newCatalogDatabaseListCommand())
	cmd.AddCommand(c.newCatalogDatabaseUpdateCommand())

	return cmd
}

func readDatabaseResourceFile(resourceFilePath string) (cmfsdk.KafkaDatabase, error) {
	data, err := os.ReadFile(resourceFilePath)
	if err != nil {
		return cmfsdk.KafkaDatabase{}, fmt.Errorf("failed to read file: %v", err)
	}

	var genericData map[string]interface{}
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &genericData)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &genericData)
	default:
		return cmfsdk.KafkaDatabase{}, errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
	}
	if err != nil {
		return cmfsdk.KafkaDatabase{}, fmt.Errorf("failed to parse input file: %w", err)
	}

	jsonBytes, err := json.Marshal(genericData)
	if err != nil {
		return cmfsdk.KafkaDatabase{}, fmt.Errorf("failed to marshal intermediate data: %w", err)
	}

	var sdkDatabase cmfsdk.KafkaDatabase
	if err = json.Unmarshal(jsonBytes, &sdkDatabase); err != nil {
		return cmfsdk.KafkaDatabase{}, fmt.Errorf("failed to bind data to KafkaDatabase model: %w", err)
	}

	return sdkDatabase, nil
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
