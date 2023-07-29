package schemaregistry

import (
	"strconv"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type getStatusOut struct {
	Name       string `human:"Name" serialized:"name"`
	State      string `human:"State" serialized:"state"`
	Offset     string `human:"Offset" serialized:"offset"`
	Timestamp  string `human:"Timestamp" serialized:"timestamp"`
	ErrorTrace string `human:"Error Trace" serialized:"error_trace"`
}

func (c *command) newExporterGetStatusCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-status <name>",
		Short: "Get the status of the schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterGetStatus,
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

func (c *command) exporterGetStatus(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	status, err := client.GetExporterStatus(args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&getStatusOut{
		Name:       status.Name,
		State:      status.State,
		Offset:     strconv.FormatInt(status.Offset, 10),
		Timestamp:  strconv.FormatInt(status.Ts, 10),
		ErrorTrace: status.Trace,
	})
	return table.Print()
}
