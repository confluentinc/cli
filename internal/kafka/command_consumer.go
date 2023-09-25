package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

type consumerCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type consumerOut struct {
	ConsumerGroupId string `human:"Consumer Group" serialized:"consumer_group_id"`
	ConsumerId      string `human:"Consumer" serialized:"consumer_id"`
	InstanceId      string `human:"Instance" serialized:"instance_id"`
	ClientId        string `human:"Client" serialized:"client_id"`
}

func newConsumerCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
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
	cmd.AddCommand(c.newGroupCommand(cfg))

	return cmd
}

func (c *consumerCommand) addConsumerGroupFlag(cmd *cobra.Command) {
	cmd.Flags().String("group", "", "Consumer group ID.")

	pcmd.RegisterFlagCompletionFunc(cmd, "group", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return pcmd.AutocompleteConsumerGroups(c.AuthenticatedCLICommand)
	})
}
