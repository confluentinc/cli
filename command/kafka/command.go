package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

type command struct {
	*cobra.Command
	config *shared.Config
}

// New returns the Cobra command for Kafka.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "kafka",
			Short: "Manage kafka.",
		},
		config: config,
	}
	err := cmd.init()
	return cmd.Command, err
}

func (c *command) init() error {
	cmdFactories := []common.CommandFactory{
		func(config *shared.Config, plugin interface{}) *cobra.Command {
			return NewClusterCommand(config, plugin.(Kafka))
		},
		func(config *shared.Config, plugin interface{}) *cobra.Command {
			return NewTopicCommand(config, plugin.(Kafka))
		},
	}
	return common.InitCommand("confluent-kafka-plugin",
		"kafka",
		c.config,
		c.Command,
		cmdFactories)
}
