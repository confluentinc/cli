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
