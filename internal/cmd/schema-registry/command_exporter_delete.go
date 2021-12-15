package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *exporterCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *exporterCommand) delete(cmd *cobra.Command, args []string) error {
	name := args[0]

	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	if _, err = srClient.DefaultApi.DeleteExporter(ctx, name); err != nil {
		return err
	}

	utils.Printf(cmd, errors.ExporterActionMsg, "Deleted", name)
	return nil
}
