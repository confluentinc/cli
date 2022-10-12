package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	ctx := dynamicconfig.NewDynamicContext(cfg.Context(), nil, nil)

	cmd := &cobra.Command{
		Use:         "stream-share",
		Aliases:     []string{"ss"},
		Short:       "Manage stream shares.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Hidden:      !cfg.IsTest && !featureflags.Manager.BoolVariation("cli.cdx", ctx, v1.CliLaunchDarklyClient, true, false),
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(c.newConsumerCommand())
	c.AddCommand(c.newProviderCommand())

	return c.Command
}
