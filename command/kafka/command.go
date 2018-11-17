package kafka

import (
	"github.com/spf13/cobra"
	"github.com/codyaray/go-printer"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/kafka"
)

var jsonPrinter = printer.NewJSONPrinter().Pretty()


type command struct {
	*cobra.Command
	config *shared.Config
}

// New returns the default command object for interacting with Kafka.
func New(config *shared.Config) (*cobra.Command, error) {
	return newCMD(config, grpcLoader)
}

// NewKafkaCommand returns a command object using a custom Kafka provider.
func NewKafkaCommand(config *shared.Config, provider common.Provider) (*cobra.Command, error) {
	return newCMD(config, provider)
}

// newCMD returns a command for interacting with Kafka.
func newCMD(config *shared.Config, plugin common.Provider) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "kafka",
			Short: "Manage kafka.",
		},
		config: config,
	}
	err := cmd.init(plugin)
	return cmd.Command, err
}

// grpcLoader is the default client loader for the CLI
func grpcLoader(i interface{}) error {
	return common.LoadPlugin(kafka.Name, i)
}

func (c *command) init(plugin common.Provider) error {
	c.AddCommand(NewClusterCommand(c.config, plugin))
	c.AddCommand(NewTopicCommand(c.config, plugin))
	c.AddCommand(NewACLCommand(c.config, plugin))

	return nil
}
