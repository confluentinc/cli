package schemaregistry

import (
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type exporterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	srClient *srsdk.APIClient
}

type exporterInfoDisplay struct {
	Name        string
	Subjects    string
	ContextType string
	Context     string
	Config      string
}

type exporterStatusDisplay struct {
	Name      string
	State     string
	Offset    string
	Timestamp string
	Trace     string
}

func newExporterCommand(prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "exporter",
		Short:       "Manage Schema Registry exporters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &exporterCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner, ExporterSubcommandFlags),
		srClient:                      srClient,
	}

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newGetConfigCommand())
	c.AddCommand(c.newGetStatusCommand())
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newPauseCommand())
	c.AddCommand(c.newResetCommand())
	c.AddCommand(c.newResumeCommand())
	c.AddCommand(c.newUpdateCommand())

	return c.Command
}
