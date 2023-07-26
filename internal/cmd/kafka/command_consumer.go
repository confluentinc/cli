package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type consumerCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type consumerOut struct {
	ConsumerGroupId string `human:"Consumer Group" serialized:"consumer_group"`
	ConsumerId      string `human:"Consumer" serialized:"consumer"`
	InstanceId      string `human:"Instance" serialized:"instance"`
	ClientId        string `human:"Client" serialized:"client"`
}

func newConsumerCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "consumer",
		Short:       "Manage Kafka consumers.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &consumerCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newListCommand())

	return cmd
}
