package schemaregistry

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/cobra"
)

func (c *exporterCommand) newResumeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "resume <name>",
		Short:       "Resume schema exporter.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.onPremResume,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *exporterCommand) onPremResume(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return resumeExporter(args[0], srClient, ctx)
}
