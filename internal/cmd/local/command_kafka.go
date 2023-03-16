package local

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

func NewKafkaCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "kafka",
			Short: "Run Kafka related commands",
			Long:  `----`,
			Args:  cobra.NoArgs,
		}, prerunner)

	c.AddCommand(c.newStartCommand())
	c.AddCommand(c.newStopCommand())
	c.AddCommand(c.newProduceCommand())
	c.AddCommand(c.newConsumeCommand())
	return c.Command
}
