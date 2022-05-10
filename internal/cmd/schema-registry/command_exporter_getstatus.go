package schemaregistry

import (
	"context"
	"strconv"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	describeStatusLabels            = []string{"Name", "State", "Offset", "Timestamp", "Trace"}
	describeStatusHumanRenames      = map[string]string{"State": "Exporter State", "Offset": "Exporter Offset", "Timestamp": "Exporter Timestamp", "Trace": "Error Trace"}
	describeStatusStructuredRenames = map[string]string{"Name": "name", "State": "state", "Offset": "offset", "Timestamp": "timestamp", "Trace": "trace"}
)

func (c *exporterCommand) newGetStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-status <name>",
		Short: "Get the status of the schema exporter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.getStatus,
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *exporterCommand) getStatus(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	return getExporterStatus(cmd, args[0], srClient, ctx)
}

func getExporterStatus(cmd *cobra.Command, name string, srClient *srsdk.APIClient, ctx context.Context) error {
	status, _, err := srClient.DefaultApi.GetExporterStatus(ctx, name)
	if err != nil {
		return err
	}

	data := &exporterStatusDisplay{
		Name:      status.Name,
		State:     status.State,
		Offset:    strconv.FormatInt(status.Offset, 10),
		Timestamp: strconv.FormatInt(status.Ts, 10),
		Trace:     status.Trace,
	}
	return output.DescribeObject(cmd, data, describeStatusLabels, describeStatusHumanRenames, describeStatusStructuredRenames)
}
