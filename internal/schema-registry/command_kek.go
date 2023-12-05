package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newKekCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "kek",
		Short:       "Manage Schema Registry Kek.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newKekCreateCommand(cfg))
	cmd.AddCommand(c.newKekDeleteCommand(cfg))
	cmd.AddCommand(c.newKekDescribeCommand(cfg))
	cmd.AddCommand(c.newKekListCommand(cfg))
	cmd.AddCommand(c.newKekUpdateCommand(cfg))

	return cmd
}
