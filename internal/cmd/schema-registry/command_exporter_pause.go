package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newExporterPauseCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause <name>",
		Short: "Pause schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterPause,
	}

	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterPause(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	if _, err := client.PauseExporter(args[0]); err != nil {
		return err
	}

	output.Printf(errors.ExporterActionMsg, "Paused", args[0])
	return nil
}
