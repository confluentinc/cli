package rtce

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

type regionCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type regionOut struct {
	ID          string `human:"ID" serialized:"id"`
	Cloud       string `human:"Cloud" serialized:"cloud"`
	Region      string `human:"Region" serialized:"region"`
	DisplayName string `human:"Display Name" serialized:"display_name"`
}

func newRegionCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command { //nolint:unparam
	cmd := &cobra.Command{
		Use:         "region",
		Short:       "Manage rtce regions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &regionCommand{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
	}

	cmd.AddCommand(
		c.newListCommand(),
	)

	return cmd
}
