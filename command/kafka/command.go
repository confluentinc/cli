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

// New returns the default command object for interacting with Kafka.
func New(config *shared.Config) (*cobra.Command, error) {
	return newCMD(config, grpcLoader)
}

// NewKafkaCommand returns a command object using a custom Kafka provider.
func NewKafkaCommand(config *shared.Config, provider func(interface{}) error) (*cobra.Command, error) {
	return newCMD(config, provider)
}

// newCMD returns a command for interacting with Kafka.
func newCMD(config *shared.Config, provider func(interface{}) error) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "kafka",
			Short: "Manage kafka.",
		},
		config: config,
	}
	err := cmd.init(provider)
	return cmd.Command, err
}

func grpcLoader(i interface{}) error {
	return common.LoadPlugin(kafka.Name, i)
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
