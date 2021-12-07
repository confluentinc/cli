package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *exporterCommand) newPauseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause <name>",
		Short: "Pause schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.pause),
	}

	output.AddFlag(cmd)

	return cmd
}

func (c *exporterCommand) pause(cmd *cobra.Command, args []string) error {
	name := args[0]

	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	if _, _, err = srClient.DefaultApi.PauseExporter(ctx, name); err != nil {
		return err
	}

	utils.Printf(cmd, errors.ExporterActionMsg, "Paused", name)
	return nil
}
