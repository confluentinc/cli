package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newDekCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "dek",
		Short:       "Manage Schema Registry dek.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newDekCreateCommand(cfg))
	// cmd.AddCommand(c.newDekGetCommand(cfg))
	// cmd.AddCommand(c.newDekDeleteCommand(cfg))
	// cmd.AddCommand(c.newDekUndeleteCommand(cfg))

	return cmd
}
