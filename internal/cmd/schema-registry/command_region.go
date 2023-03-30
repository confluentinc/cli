package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type schemaRegistryCloudRegionHumanOut struct {
	ID         string `human:"ID"`
	Name       string `human:"Name"`
	Cloud      string `human:"Cloud"`
	RegionName string `human:"Region Name"`
	Packages   string `human:"Packages"`
}

type schemaRegistryCloudRegionSerializedOut struct {
	ID         string   `serialized:"id"`
	Name       string   `serialized:"name"`
	Cloud      string   `serialized:"cloud"`
	RegionName string   `serialized:"region_name"`
	Packages   []string `serialized:"packages"`
}

func (c *command) newRegionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "region",
		Short:       "Manage Schema Registry cloud regions.",
		Long:        "Use this command to manage Schema Registry cloud regions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newListCommand())

	return cmd
}
