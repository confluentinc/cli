package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "flink",
		Short:       "Manage Apache Flink.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newComputePoolCommand(cfg))
	cmd.AddCommand(c.newRegionCommand())
	cmd.AddCommand(c.newShellCommand(cfg, prerunner))
	cmd.AddCommand(c.newStatementCommand())

	dc := dynamicconfig.New(cfg, nil)
	_ = dc.ParseFlagsIntoConfig(cmd)
	if cfg.IsTest || featureflags.Manager.BoolVariation("cli.flink.open_preview", dc.Context(), config.CliLaunchDarklyClient, true, false) {
		cmd.AddCommand(c.newIamBindingCommand())
	}

	return cmd
}
