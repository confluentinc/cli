package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type catalogOut struct {
	CreationTime string   `human:"Creation Time" serialized:"creation_time"`
	Name         string   `human:"Name" serialized:"name"`
	Databases    []string `human:"Databases" serialized:"databases"`
}

// localCatalog is a local struct with YAML tags that matches the SDK KafkaCatalog structure
type localCatalog struct {
	ApiVersion string                  `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                  `yaml:"kind" json:"kind"`
	Metadata   map[string]interface{}  `yaml:"metadata" json:"metadata"`
	Spec       map[string]interface{}  `yaml:"spec" json:"spec"`
	Status     *map[string]interface{} `yaml:"status,omitempty" json:"status,omitempty"`
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
