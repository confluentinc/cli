package schemaregistry

import (
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type exporterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	srClient *srsdk.APIClient
}

type exporterInfoDisplay struct {
	Name          string
	Subjects      string
	SubjectFormat string
	ContextType   string
	Context       string
	Config        string
}

type exporterStatusDisplay struct {
	Name      string
	State     string
	Offset    string
	Timestamp string
	Trace     string
}

func newExporterCommand(cfg *v1.Config, prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "exporter",
		Short:       "Manage Schema Registry exporters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &exporterCommand{srClient: srClient}

	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)
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
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
		c.AddCommand(c.newCreateCommandOnPrem())
		c.AddCommand(c.newDeleteCommandOnPrem())
		c.AddCommand(c.newDescribeCommandOnPrem())
		c.AddCommand(c.newGetConfigCommandOnPrem())
		c.AddCommand(c.newGetStatusCommandOnPrem())
		c.AddCommand(c.newListCommandOnPrem())
		c.AddCommand(c.newPauseCommandOnPrem())
		c.AddCommand(c.newResetCommandOnPrem())
		c.AddCommand(c.newResumeCommandOnPrem())
		c.AddCommand(c.newUpdateCommandOnPrem())
	}

	return c.Command
}

func addContextTypeFlag(cmd *cobra.Command) {
	cmd.Flags().String("context-type", "AUTO", `Exporter context type. One of "AUTO", "CUSTOM" or "NONE".`)
	pcmd.RegisterFlagCompletionFunc(cmd, "context-type", func(_ *cobra.Command, _ []string) []string { return []string{"AUTO", "CUSTOM", "NONE"} })
}
