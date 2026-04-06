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

type secretMappingOut struct {
	CreationTime string `human:"Creation Time" serialized:"creation_time"`
	Name         string `human:"Name" serialized:"name"`
	SecretName   string `human:"Secret Name" serialized:"secret_name"`
}

func (c *command) newSecretMappingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "secret-mapping",
		Short:       "Manage Flink secret mappings.",
		Long:        "Manage Flink environment secret mappings in Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newSecretMappingCreateCommand())
	cmd.AddCommand(c.newSecretMappingDeleteCommand())
	cmd.AddCommand(c.newSecretMappingDescribeCommand())
	cmd.AddCommand(c.newSecretMappingListCommand())
	cmd.AddCommand(c.newSecretMappingUpdateCommand())

	return cmd
}

func printSecretMappingOutput(cmd *cobra.Command, sdkMapping cmfsdk.EnvironmentSecretMapping) error {
	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		var creationTime, name, secretName string
		if sdkMapping.Metadata != nil {
			if sdkMapping.Metadata.CreationTimestamp != nil {
				creationTime = *sdkMapping.Metadata.CreationTimestamp
			}
			if sdkMapping.Metadata.Name != nil {
				name = *sdkMapping.Metadata.Name
			}
		}
		if sdkMapping.Spec != nil {
			secretName = sdkMapping.Spec.SecretName
		}
		table.Add(&secretMappingOut{
			CreationTime: creationTime,
			Name:         name,
			SecretName:   secretName,
		})
		return table.Print()
	}

	localMapping := convertSdkSecretMappingToLocalSecretMapping(sdkMapping)
	return output.SerializedOutput(cmd, localMapping)
}

func readSecretMappingResourceFile(resourceFilePath string) (cmfsdk.EnvironmentSecretMapping, error) {
	data, err := os.ReadFile(resourceFilePath)
	if err != nil {
		return cmfsdk.EnvironmentSecretMapping{}, fmt.Errorf("failed to read file: %w", err)
	}

	var genericData map[string]interface{}
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &genericData)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &genericData)
	default:
		return cmfsdk.EnvironmentSecretMapping{}, errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
	}
	if err != nil {
		return cmfsdk.EnvironmentSecretMapping{}, fmt.Errorf("failed to parse input file: %w", err)
	}

	jsonBytes, err := json.Marshal(genericData)
	if err != nil {
		return cmfsdk.EnvironmentSecretMapping{}, fmt.Errorf("failed to marshal intermediate data: %w", err)
	}

	var sdkMapping cmfsdk.EnvironmentSecretMapping
	if err = json.Unmarshal(jsonBytes, &sdkMapping); err != nil {
		return cmfsdk.EnvironmentSecretMapping{}, fmt.Errorf("failed to bind data to EnvironmentSecretMapping model: %w", err)
	}

	return sdkMapping, nil
}

func convertSdkSecretMappingToLocalSecretMapping(sdkMapping cmfsdk.EnvironmentSecretMapping) LocalSecretMapping {
	localMapping := LocalSecretMapping{
		ApiVersion: sdkMapping.ApiVersion,
		Kind:       sdkMapping.Kind,
	}

	if sdkMapping.Metadata != nil {
		localMapping.Metadata = LocalSecretMappingMetadata{
			Name:              sdkMapping.Metadata.GetName(),
			CreationTimestamp: sdkMapping.Metadata.CreationTimestamp,
			UpdateTimestamp:   sdkMapping.Metadata.UpdateTimestamp,
			Uid:               sdkMapping.Metadata.Uid,
			Labels:            sdkMapping.Metadata.Labels,
			Annotations:       sdkMapping.Metadata.Annotations,
		}
	}

	if sdkMapping.Spec != nil {
		localMapping.Spec = LocalSecretMappingSpec{
			SecretName: sdkMapping.Spec.SecretName,
		}
	}

	return localMapping
}
