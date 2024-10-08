package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

type clientConfigCommand struct {
	*pcmd.AuthenticatedCLICommand
	clientId string
}

func newClientConfigCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "client-config",
		Short:       "Manage Kafka Clients configuration files.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &clientConfigCommand{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
		clientId:                cfg.Version.ClientID,
	}

	cmd.AddCommand(c.newCreateCommand())

	return cmd
}
