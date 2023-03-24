package local

import (
	"github.com/spf13/cobra"
)

const imageName = "523370736235.dkr.ecr.us-west-2.amazonaws.com/confluentinc/kafka-local:latest"
const localhostPrefix = "http://localhost:%s"

func (c *command) newKafkaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Manage a single-node instance of Apache Kafka.",
	}

	cmd.AddCommand(c.newStartCommand())
	cmd.AddCommand(c.newStopCommand())
	cmd.AddCommand(c.newTopicCommand())

	return cmd
}
