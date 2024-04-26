package schemaregistry

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newExporterStatusCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Manage the schema exporter status.",
	}

	cmd.AddCommand(c.newExporterStatusDescribeCommand(cfg))

	return cmd
}
