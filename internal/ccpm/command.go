package ccpm

import (
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "ccpm",
		Short:       "Manage Custom Connect Plugin Management (CCPM).",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Hidden:      !(cfg.IsTest || featureflags.Manager.BoolVariation("custom-connect.plugin.enabled", cfg.Context(), config.CliLaunchDarklyClient, true, false)),
	}

	cmd.AddCommand(newPluginCommand(prerunner))
	return cmd
}
