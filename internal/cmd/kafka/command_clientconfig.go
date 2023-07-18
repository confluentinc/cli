package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type clientConfigCommand struct {
	*pcmd.HasAPIKeyCLICommand
}

func newClientConfigCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "client-config",
		Short:       "Manage Kafka Clients configuration files.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &clientConfigCommand{pcmd.NewHasAPIKeyCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand(cfg, prerunner))

	return cmd
}
