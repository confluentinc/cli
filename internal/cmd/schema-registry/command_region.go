package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type schemaRegistryCloudRegion struct {
	ID          string   `human:"ID" serialized:"id"`
	Cloud       string   `human:"Cloud" serialized:"cloud"`
	RegionName  string   `human:"Region Name" serialized:"region_name"`
	DisplayName string   `human:"Display Name" serialized:"display_name"`
	Packages    []string `human:"Packages" serialized:"packages"`
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
