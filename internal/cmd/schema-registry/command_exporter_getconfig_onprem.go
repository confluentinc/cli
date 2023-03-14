package schemaregistry

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (c *command) newExporterGetConfigCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "get-config <name>",
		Short:       "Get the configurations of the schema exporter.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.exporterGetConfigOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlagWithDefaultValue(cmd, output.JSON.String())

	return cmd
}

func (c *command) exporterGetConfigOnPrem(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return getExporterConfig(cmd, args[0], srClient, ctx)
}
