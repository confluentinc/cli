package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

type accessPointCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newAccessPointCommand(prerunner pcmd.PreRunner, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "access-point",
		Short:       "Manage access points.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &accessPointCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newPrivateLinkCommand(cfg))
	cmd.AddCommand(c.newPrivateNetworkInterfaceCommand())

	return cmd
}
