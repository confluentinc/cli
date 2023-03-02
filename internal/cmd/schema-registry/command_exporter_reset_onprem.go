package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *exporterCommand) newResetCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "reset <name>",
		Short:       "Reset schema exporter.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.onPremReset,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *exporterCommand) onPremReset(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return resetExporter(args[0], srClient, ctx)
}
