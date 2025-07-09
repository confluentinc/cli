package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

type consumerCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type consumerOut struct {
	ConsumerGroup string `human:"Consumer Group" serialized:"consumer_group"`
	Consumer      string `human:"Consumer" serialized:"consumer"`
	Instance      string `human:"Instance" serialized:"instance"`
	Client        string `human:"Client" serialized:"client"`
}

func newConsumerCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumer",
		Short: "Manage Kafka consumers.",
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

		return pcmd.AutocompleteConsumerGroups(cmd, c.AuthenticatedCLICommand)
	})
}
