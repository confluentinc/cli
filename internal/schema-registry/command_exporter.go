package schemaregistry

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

const exporterActionMsg = "%s schema exporter \"%s\".\n"

func (c *command) newExporterCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exporter",
		Short: "Manage Schema Registry exporters.",
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

func addContextTypeFlag(cloud bool, cmd *cobra.Command) {
	arr := []string{"auto", "custom", "none"}
	if cloud {
		arr = append(arr, "default")
		slices.Sort(arr)
	}
	cmd.Flags().String("context-type", arr[0], fmt.Sprintf(`Exporter context type. One of %s.`, utils.ArrayToCommaDelimitedString(arr, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "context-type", func(_ *cobra.Command, _ []string) []string { return arr })
}
