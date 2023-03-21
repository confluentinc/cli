package local

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *kafkaCommand) newTopicCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "Run Kafka topic related commands",
		Long:  `Run Kafka commands including produce/consume and list topics`,
		Args:  cobra.NoArgs,
	}

	// c := &kafkaCommand{pcmd.NewAnonymousCLICommand(cmd, prerunner)}

	// cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newProduceCommand())
	cmd.AddCommand(c.newConsumeCommand())
	return cmd
}
