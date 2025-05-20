package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type catalogOut struct {
	Name         string   `human:"Name" serialized:"name"`
	CreationTime string   `human:"CreationTime" serialized:"creation_time"`
	Databases    []string `human:"Databases" serialized:"databases"`
}

func (c *command) newCatalogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "catalog",
		Short:       "Manage Flink Catalog in Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newCatalogCreateCommand())
	cmd.AddCommand(c.newCatalogDeleteCommand())
	cmd.AddCommand(c.newCatalogDescribeCommand())
	cmd.AddCommand(c.newCatalogListCommand())

	return cmd
}
