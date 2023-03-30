package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *command) newExporterGetStatusCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "get-status <name>",
		Short:       "Get the status of the schema exporter.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.exporterGetStatusOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterGetStatusOnPrem(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return getExporterStatus(cmd, args[0], srClient, ctx)
}
