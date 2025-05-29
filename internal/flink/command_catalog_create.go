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

	var catalog cmfsdk.KafkaCatalog
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &catalog)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &catalog)
	default:
		return errors.NewErrorWithSuggestions(fmt.Sprintf("unsupported file format: %s", ext), "Supported file formats are .json, .yaml, and .yml.")
	}
	if err != nil {
		return err
	}

	outputCatalog, err := client.CreateCatalog(c.createContext(), catalog)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)

		// Populate the databases field with the names of the databases
		databases := make([]string, 0, len(outputCatalog.GetSpec().KafkaClusters))
		for _, kafkaCluster := range outputCatalog.GetSpec().KafkaClusters {
			databases = append(databases, kafkaCluster.DatabaseName)
		}

		// nil pointer handling for creation timestamp
		var creationTime string
		if outputCatalog.GetMetadata().CreationTimestamp != nil {
			creationTime = *outputCatalog.GetMetadata().CreationTimestamp
		} else {
			creationTime = ""
		}

		table.Add(&catalogOut{
			CreationTime: creationTime,
			Name:         outputCatalog.GetMetadata().Name,
			Databases:    databases,
		})
		return table.Print()
	}

	return output.SerializedOutput(cmd, outputCatalog)
}
