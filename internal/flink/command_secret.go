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

type secretOut struct {
	CreationTime string `human:"Creation Time" serialized:"creation_time"`
	Name         string `human:"Name" serialized:"name"`
}

func (c *command) newSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "secret",
		Short:       "Manage Flink secrets in Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newSecretCreateCommand())
	cmd.AddCommand(c.newSecretDeleteCommand())
	cmd.AddCommand(c.newSecretDescribeCommand())
	cmd.AddCommand(c.newSecretListCommand())
	cmd.AddCommand(c.newSecretUpdateCommand())

	return cmd
}

func printSecretOutput(cmd *cobra.Command, sdkSecret cmfsdk.Secret) error {
	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		var creationTime string
		if sdkSecret.Metadata.CreationTimestamp != nil {
			creationTime = *sdkSecret.Metadata.CreationTimestamp
		}
		table.Add(&secretOut{
			CreationTime: creationTime,
			Name:         sdkSecret.Metadata.Name,
		})
		return table.Print()
	}

	localSecret := convertSdkSecretToLocalSecret(sdkSecret)
	return output.SerializedOutput(cmd, localSecret)
}

func readSecretResourceFile(resourceFilePath string) (cmfsdk.Secret, error) {
	data, err := os.ReadFile(resourceFilePath)
	if err != nil {
		return cmfsdk.Secret{}, fmt.Errorf("failed to read file: %w", err)
	}

	var genericData map[string]interface{}
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &genericData)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &genericData)
	default:
		return cmfsdk.Secret{}, errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
	}
	if err != nil {
		return cmfsdk.Secret{}, fmt.Errorf("failed to parse input file: %w", err)
	}

	jsonBytes, err := json.Marshal(genericData)
	if err != nil {
		return cmfsdk.Secret{}, fmt.Errorf("failed to marshal intermediate data: %w", err)
	}

	var sdkSecret cmfsdk.Secret
	if err = json.Unmarshal(jsonBytes, &sdkSecret); err != nil {
		return cmfsdk.Secret{}, fmt.Errorf("failed to bind data to Secret model: %w", err)
	}

	return sdkSecret, nil
}

func convertSdkSecretToLocalSecret(sdkSecret cmfsdk.Secret) LocalSecret {
	localSecret := LocalSecret{
		ApiVersion: sdkSecret.ApiVersion,
		Kind:       sdkSecret.Kind,
		Metadata: LocalSecretMetadata{
			Name:              sdkSecret.Metadata.Name,
			CreationTimestamp: sdkSecret.Metadata.CreationTimestamp,
			UpdateTimestamp:   sdkSecret.Metadata.UpdateTimestamp,
			Uid:               sdkSecret.Metadata.Uid,
			Labels:            sdkSecret.Metadata.Labels,
			Annotations:       sdkSecret.Metadata.Annotations,
		},
		Spec: LocalSecretSpec{
			Data: sdkSecret.Spec.Data,
		},
	}

	if sdkSecret.Status != nil {
		localSecret.Status = &LocalSecretStatus{
			Version:      sdkSecret.Status.Version,
			Environments: sdkSecret.Status.Environments,
		}
	}

	return localSecret
}
