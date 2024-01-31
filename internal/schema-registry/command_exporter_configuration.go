package schemaregistry

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newExporterConfigurationCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Manage the schema exporter configuration.",
	}

	cmd.AddCommand(c.newExporterConfigurationDescribeCommand(cfg))

	return cmd
}
