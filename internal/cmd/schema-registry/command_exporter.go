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

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newExporterCreateCommand())
		cmd.AddCommand(c.newExporterDeleteCommand())
		cmd.AddCommand(c.newExporterDescribeCommand())
		cmd.AddCommand(c.newExporterGetConfigCommand())
		cmd.AddCommand(c.newExporterGetStatusCommand())
		cmd.AddCommand(c.newExporterListCommand())
		cmd.AddCommand(c.newExporterPauseCommand())
		cmd.AddCommand(c.newExporterResetCommand())
		cmd.AddCommand(c.newExporterResumeCommand())
		cmd.AddCommand(c.newExporterUpdateCommand())
	} else {
		cmd.AddCommand(c.newExporterCreateCommandOnPrem())
		cmd.AddCommand(c.newExporterDeleteCommandOnPrem())
		cmd.AddCommand(c.newExporterDescribeCommandOnPrem())
		cmd.AddCommand(c.newExporterGetConfigCommandOnPrem())
		cmd.AddCommand(c.newExporterGetStatusCommandOnPrem())
		cmd.AddCommand(c.newExporterListCommandOnPrem())
		cmd.AddCommand(c.newExporterPauseCommandOnPrem())
		cmd.AddCommand(c.newExporterResetCommandOnPrem())
		cmd.AddCommand(c.newExporterResumeCommandOnPrem())
		cmd.AddCommand(c.newExporterUpdateCommandOnPrem())
	}

	return cmd
}

func addContextTypeFlag(cmd *cobra.Command) {
	cmd.Flags().String("context-type", "AUTO", `Exporter context type. One of "AUTO", "CUSTOM" or "NONE".`)
	pcmd.RegisterFlagCompletionFunc(cmd, "context-type", func(_ *cobra.Command, _ []string) []string { return []string{"AUTO", "CUSTOM", "NONE"} })
}
