package schemaregistry

import (
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type exporterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	srClient *srsdk.APIClient
}

func newExporterCommand(cfg *v1.Config, prerunner pcmd.PreRunner, srClient *srsdk.APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "exporter",
		Short:       "Manage Schema Registry exporters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	c := &exporterCommand{srClient: srClient}

	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)
		cmd.AddCommand(c.newCreateCommand())
		cmd.AddCommand(c.newDeleteCommand())
		cmd.AddCommand(c.newDescribeCommand())
		cmd.AddCommand(c.newGetConfigCommand())
		cmd.AddCommand(c.newGetStatusCommand())
		cmd.AddCommand(c.newListCommand())
		cmd.AddCommand(c.newPauseCommand())
		cmd.AddCommand(c.newResetCommand())
		cmd.AddCommand(c.newResumeCommand())
		cmd.AddCommand(c.newUpdateCommand())
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
		cmd.AddCommand(c.newCreateCommandOnPrem())
		cmd.AddCommand(c.newDeleteCommandOnPrem())
		cmd.AddCommand(c.newDescribeCommandOnPrem())
		cmd.AddCommand(c.newGetConfigCommandOnPrem())
		cmd.AddCommand(c.newGetStatusCommandOnPrem())
		cmd.AddCommand(c.newListCommandOnPrem())
		cmd.AddCommand(c.newPauseCommandOnPrem())
		cmd.AddCommand(c.newResetCommandOnPrem())
		cmd.AddCommand(c.newResumeCommandOnPrem())
		cmd.AddCommand(c.newUpdateCommandOnPrem())
	}

	return cmd
}

func addContextTypeFlag(cmd *cobra.Command) {
	cmd.Flags().String("context-type", "AUTO", `Exporter context type. One of "AUTO", "CUSTOM" or "NONE".`)
	pcmd.RegisterFlagCompletionFunc(cmd, "context-type", func(_ *cobra.Command, _ []string) []string { return []string{"AUTO", "CUSTOM", "NONE"} })
}
