package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

const exporterActionMsg = "%s schema exporter \"%s\".\n"

func (c *command) newExporterCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "exporter",
		Short:       "Manage Schema Registry exporters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newExporterCreateCommand(cfg))
	cmd.AddCommand(c.newExporterDeleteCommand(cfg))
	cmd.AddCommand(c.newExporterConfigurationCommand(cfg))
	cmd.AddCommand(c.newExporterDescribeCommand(cfg))
	cmd.AddCommand(c.newExporterListCommand(cfg))
	cmd.AddCommand(c.newExporterPauseCommand(cfg))
	cmd.AddCommand(c.newExporterResetCommand(cfg))
	cmd.AddCommand(c.newExporterResumeCommand(cfg))
	cmd.AddCommand(c.newExporterStatusCommand(cfg))
	cmd.AddCommand(c.newExporterUpdateCommand(cfg))

	return cmd
}

func addContextTypeFlag(cmd *cobra.Command) {
	arr := []string{"auto", "custom", "none"}
	cmd.Flags().String("context-type", arr[0], fmt.Sprintf(`Exporter context type. One of %s.`, utils.ArrayToCommaDelimitedString(arr, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "context-type", func(_ *cobra.Command, _ []string) []string { return arr })
}
