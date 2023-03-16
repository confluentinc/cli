package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type clientConfigCommand struct {
	*pcmd.HasAPIKeyCLICommand
}

func newClientConfigCommand(prerunner pcmd.PreRunner, clientId string) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "client-config",
		Short:       "Manage Kafka Clients configuration files.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &clientConfigCommand{pcmd.NewHasAPIKeyCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand(prerunner, clientId))

	return cmd
}
