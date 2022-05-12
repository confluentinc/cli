package schemaregistry

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/cobra"
)

func (c *exporterCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe <name>",
		Short:       "Describe the schema exporter.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.onPremDescribe,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *exporterCommand) onPremDescribe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return describeExporter(cmd, args[0], srClient, ctx)
}
