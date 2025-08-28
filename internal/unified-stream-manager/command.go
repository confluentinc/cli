package unifiedstreammanager

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(prerunner pcmd.PreRunner, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "unified-stream-manager",
		Aliases:     []string{"usm"},
		Short:       "Manage Unified Stream Manager clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Hidden:      !(cfg.IsTest || featureflags.Manager.BoolVariation("cli.usm", cfg.Context(), config.CliLaunchDarklyClient, true, false)),
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newKafkaCommand())
	cmd.AddCommand(c.newConnectCommand())

	return cmd
}

func (c *command) addRegionFlag(cmd *cobra.Command) {
	cmd.Flags().String("region", "", "Specify the cloud region.")
}
