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

func (c *command) newCatalogCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <resourceFilePath>",
		Short: "Create a Flink catalog.",
		Long:  "Create a Flink catalog in Confluent Platform that provides metadata about tables and other database objects such as views and functions.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.catalogCreate,
	}

	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) catalogCreate(cmd *cobra.Command, args []string) error {
	resourceFilePath := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	// Read file contents
	data, err := os.ReadFile(resourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	var genericData map[string]interface{}
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &genericData)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &genericData)
	default:
		return errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
	}
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(genericData)
	if err != nil {
		return fmt.Errorf("failed to marshal intermediate data: %w", err)
	}

	var sdkCatalog cmfsdk.KafkaCatalog
	if err = json.Unmarshal(jsonBytes, &sdkCatalog); err != nil {
		return fmt.Errorf("failed to bind data to KafkaCatalog model: %w", err)
	}

	sdkOutputCatalog, err := client.CreateCatalog(c.createContext(), sdkCatalog)
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
