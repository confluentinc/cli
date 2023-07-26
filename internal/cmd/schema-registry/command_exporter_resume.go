package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newExporterResumeCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <name>",
		Short: "Resume schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterResume,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
	pcmd.AddOutputFlag(cmd)

	if cfg.IsCloudLogin() {
		// Deprecated
		pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-key"))

		// Deprecated
		pcmd.AddApiSecretFlag(cmd)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-secret"))
	}

	return cmd
}

func (c *command) exporterResume(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	if _, err := client.ResumeExporter(args[0]); err != nil {
		return err
	}

	output.Printf(errors.ExporterActionMsg, "Resumed", args[0])
	return nil
}
