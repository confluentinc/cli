package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	dynamicCtx := dynamicconfig.NewDynamicContext(cfg.Context(), nil, nil)

	cmd := &cobra.Command{
		Use:         "stream-share",
		Aliases:     []string{"ss"},
		Short:       "Manage stream shares.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	if !featureflags.Manager.BoolVariation("cli.cdx", dynamicCtx, v1.CliLaunchDarklyClient, true, false) {
		cmd.Hidden = true
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	c.AddCommand(newProviderCommand(prerunner))

	return c.Command
}
