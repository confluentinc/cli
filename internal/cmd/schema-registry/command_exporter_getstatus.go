package schemaregistry

import (
	"context"
	"strconv"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type getStatusOut struct {
	Name      string `human:"Name" serialized:"name"`
	State     string `human:"Exporter State" serialized:"state"`
	Offset    string `human:"Exporter Offset" serialized:"offset"`
	Timestamp string `human:"Exporter Timestamp" serialized:"timestamp"`
	Trace     string `human:"Error Trace" serialized:"trace"`
}

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

	table := output.NewTable(cmd)
	table.Add(&getStatusOut{
		Name:      status.Name,
		State:     status.State,
		Offset:    strconv.FormatInt(status.Offset, 10),
		Timestamp: strconv.FormatInt(status.Ts, 10),
		Trace:     status.Trace,
	})
	return table.Print()
}
