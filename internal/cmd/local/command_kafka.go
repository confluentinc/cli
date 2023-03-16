package local

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type kafkaCommand struct {
	*pcmd.CLICommand
}

func NewKafkaCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Run Kafka related commands",
		Long:  `----`,
		Args:  cobra.NoArgs,
	}

	c := &kafkaCommand{pcmd.NewAnonymousCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newStartCommand())
	cmd.AddCommand(c.newStopCommand())
	cmd.AddCommand(c.newProduceCommand())
	cmd.AddCommand(c.newConsumeCommand())
	return cmd
}
