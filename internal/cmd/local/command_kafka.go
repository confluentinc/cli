package local

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type kafkaCommand struct {
	*pcmd.CLICommand
}

func newKafkaCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Run Kafka related commands",
		Long:  `Run Kafka commands including starting kafka service, produce/consume and stopping the service`,
		Args:  cobra.NoArgs,
	}

	c := &kafkaCommand{pcmd.NewAnonymousCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newStartCommand())
	cmd.AddCommand(c.newStopCommand())
	cmd.AddCommand(c.newTopicCommand(prerunner))
	return cmd
}
