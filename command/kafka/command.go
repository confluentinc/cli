package kafka

import (
	"github.com/codyaray/go-printer"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/kafka"
	"github.com/spf13/cobra"
)

var jsonPrinter = printer.NewJSONPrinter().Pretty()

// Client handles communication with the service API
var Client kafka.Kafka

type command struct {
	*cobra.Command
	config *shared.Config
}

// New returns the Cobra command for Kafka.
func New(config *shared.Config, run func(interface{}) error) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "kafka",
			Short: "Manage kafka.",
		},
		config: config,
	}
	err := cmd.init(run)
	return cmd.Command, err
}

func (c *command) init(run func(interface{}) error) error {
	// All commands require login first
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := c.config.CheckLogin(); err != nil {
			return common.HandleError(err, cmd)
		}
		// Lazy load plugin to avoid unnecessarily spawning child processes
		return run(&Client)
	}

	c.AddCommand(NewClusterCommand(c.config))
	c.AddCommand(NewTopicCommand(c.config))
	c.AddCommand(NewACLCommand(c.config))

	return nil
}
