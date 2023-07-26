package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
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

func newConsumerCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "consumer",
		Short:       "Manage Kafka consumers.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &consumerCommand{}

	if cfg.IsCloudLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)

		cmd.AddCommand(c.newListCommand())
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
		c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

		cmd.AddCommand(c.newListCommandOnPrem())
	}

	return cmd
}
