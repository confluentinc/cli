package connect

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

type command struct {
	*cobra.Command
	config  *shared.Config
	connect Connect
}

// New returns the Cobra command for Connect.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "connect",
			Short: "Manage connect.",
		},
		config: config,
	}
	err := cmd.init()
	return cmd.Command, err
}

func (c *command) init() error {
	cmdFactories := []common.CommandFactory{
		func(config *shared.Config, plugin interface{}) *cobra.Command {
			return NewSink(config, plugin.(Connect))
		},
	}
	return common.InitCommand("confluent-connect-plugin",
		"connect",
		c.config,
		c.Command,
		cmdFactories)
}
