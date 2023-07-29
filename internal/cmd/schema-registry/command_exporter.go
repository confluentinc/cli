package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func (c *command) newExporterCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "exporter",
		Short:       "Manage Schema Registry exporters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newExporterCreateCommand(cfg))
	cmd.AddCommand(c.newExporterDeleteCommand(cfg))
	cmd.AddCommand(c.newExporterDescribeCommand(cfg))
	cmd.AddCommand(c.newExporterGetConfigCommand(cfg))
	cmd.AddCommand(c.newExporterGetStatusCommand(cfg))
	cmd.AddCommand(c.newExporterListCommand(cfg))
	cmd.AddCommand(c.newExporterPauseCommand(cfg))
	cmd.AddCommand(c.newExporterResetCommand(cfg))
	cmd.AddCommand(c.newExporterResumeCommand(cfg))
	cmd.AddCommand(c.newExporterUpdateCommand(cfg))

	return cmd
}

func addContextTypeFlag(cmd *cobra.Command) {
	cmd.Flags().String("context-type", "AUTO", `Exporter context type. One of "AUTO", "CUSTOM" or "NONE".`)
	pcmd.RegisterFlagCompletionFunc(cmd, "context-type", func(_ *cobra.Command, _ []string) []string { return []string{"AUTO", "CUSTOM", "NONE"} })
}
