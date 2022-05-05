package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *exporterCommand) newResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <name>",
		Short: "Resume schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.resume,
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *exporterCommand) resume(cmd *cobra.Command, args []string) error {
	name := args[0]

	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	if _, _, err = srClient.DefaultApi.ResumeExporter(ctx, name); err != nil {
		return err
	}

	utils.Printf(cmd, errors.ExporterActionMsg, "Resumed", name)
	return nil
}
