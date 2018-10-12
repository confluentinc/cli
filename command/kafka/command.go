package kafka

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

type command struct {
	*cobra.Command
	config *shared.Config
	kafka  Kafka
}

// New returns the Cobra command for Kafka.
func New(config *shared.Config, kafka Kafka) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "kafka",
			Short: "Manage kafka.",
		},
		config: config,
		kafka: kafka,
	}
	err := cmd.init()
	return cmd.Command, err
}

func (c *command) init() error {
	// remove redundant help command
	c.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	// All commands require login first
	c.Command.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if err := c.config.CheckLogin(); err != nil {
			_ = common.HandleError(err, cmd)
			os.Exit(1)
		}
	}

	c.AddCommand(NewClusterCommand(c.config, c.kafka))
	c.AddCommand(NewTopicCommand(c.config, c.kafka))
	c.AddCommand(NewAclCommand(c.config, c.kafka))
	return nil
}
