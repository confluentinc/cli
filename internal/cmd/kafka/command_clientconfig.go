package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type clientConfigCommand struct {
	*pcmd.HasAPIKeyCLICommand
	prerunner       pcmd.PreRunner
	clientId        string
	analyticsClient analytics.Client
}

func newClientConfigCommand(prerunner pcmd.PreRunner, clientID string, analyticsClient analytics.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "client-config",
		Short:       "Manage Kafka Clients configuration files.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &clientConfigCommand{
		HasAPIKeyCLICommand: pcmd.NewHasAPIKeyCLICommand(cmd, prerunner),
		prerunner:           prerunner,
		clientId:            clientID,
		analyticsClient:     analyticsClient,
	}

	c.AddCommand(c.newCreateCommand())

	return c.Command
}
