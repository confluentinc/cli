package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type regionCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type schemaRegistryCloudRegion struct {
	ID          string   `json:"id" hcl:"id"`
	Cloud       string   `json:"cloud" hcl:"cloud"`
	RegionName  string   `json:"region_name" hcl:"region_name"`
	DisplayName string   `json:"display_name" hcl:"display_name"`
	Packages    []string `json:"packages" hcl:"packages"`
}

func newRegionCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "region",
		Short:       "Manage Schema Registry cloud regions.",
		Long:        "Use this command to manage Schema Registry cloud regions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &regionCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	c.AddCommand(c.newListCommand())

	return c.Command
}
