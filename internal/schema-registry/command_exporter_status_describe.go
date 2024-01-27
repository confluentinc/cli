package schemaregistry

import (
	"strconv"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type statusOut struct {
	Name       string `human:"Name" serialized:"name"`
	State      string `human:"State" serialized:"state"`
	Offset     string `human:"Offset" serialized:"offset"`
	Timestamp  string `human:"Timestamp" serialized:"timestamp"`
	ErrorTrace string `human:"Error Trace" serialized:"error_trace"`
}

func (c *command) newExporterStatusDescribeCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe the schema exporter status.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.exporterStatusDescribe,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) exporterStatusDescribe(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	status, err := client.GetExporterStatus(args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&statusOut{
		Name:       status.GetName(),
		State:      status.GetState(),
		Offset:     strconv.FormatInt(status.GetOffset(), 10),
		Timestamp:  strconv.FormatInt(status.GetTs(), 10),
		ErrorTrace: status.GetTrace(),
	})
	return table.Print()
}
