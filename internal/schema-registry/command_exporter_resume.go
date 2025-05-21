package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newExporterResumeCommand(cfg *config.Config) *cobra.Command {
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
		addCaLocationAndClientPathFlags(cmd)
	}
	addSchemaRegistryEndpointFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterResume(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	if _, err := client.ResumeExporter(args[0]); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, exporterActionMsg, "Resumed", args[0])
	return nil
}
