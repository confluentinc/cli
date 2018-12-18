package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

// New returns the Cobra command for Kafka.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Manage kafka.",
	}

	cmdFactories := []common.CommandFactory{
		func(config *shared.Config, plugin interface{}) *cobra.Command {
			return NewClusterCommand(config, plugin.(Kafka))
		},
		func(config *shared.Config, plugin interface{}) *cobra.Command {
			return NewTopicCommand(config, plugin.(Kafka))
		},
	}
	err := common.InitCommand("confluent-kafka-plugin",
		"kafka",
		config,
		cmd,
		cmdFactories)

	return cmd, err
}
