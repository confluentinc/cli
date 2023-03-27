package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func (c *command) newSchemaCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "schema",
		Short:       "Manage Schema Registry schemas.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newSchemaCreateCommand())
		cmd.AddCommand(c.newSchemaDeleteCommand())
		cmd.AddCommand(c.newSchemaDescribeCommand())
		cmd.AddCommand(c.newSchemaListCommand())
	} else {
		cmd.AddCommand(c.newSchemaCreateCommandOnPrem())
		cmd.AddCommand(c.newSchemaDeleteCommandOnPrem())
		cmd.AddCommand(c.newSchemaDescribeCommandOnPrem())
		cmd.AddCommand(c.newSchemaListCommandOnPrem())
	}

	return cmd
}
