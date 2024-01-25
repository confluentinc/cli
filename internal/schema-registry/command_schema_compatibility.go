package schemaregistry

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newSchemaCompatibilityCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compatibility",
		Short: "Validate schema compatibility.",
	}

	cmd.AddCommand(c.newSchemaCompatibilityValidateCommand(cfg))

	return cmd
}
