package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type listOut struct {
	Exporter string `human:"Exporter" serialized:"exporter"`
}

func (c *command) newExporterListCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all schema exporters.",
		Args:  cobra.NoArgs,
		RunE:  c.exporterList,
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

func (c *command) exporterList(cmd *cobra.Command, _ []string) error {
	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	exporters, err := client.GetExporters()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, exporter := range exporters {
		list.Add(&listOut{Exporter: exporter})
	}
	return list.Print()
}
