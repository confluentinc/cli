package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newExporterResetCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset <name>",
		Short: "Reset schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterReset,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterReset(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	if _, err := client.ResetExporter(args[0]); err != nil {
		return err
	}

	output.Printf(errors.ExporterActionMsg, "Reset", args[0])
	return nil
}
