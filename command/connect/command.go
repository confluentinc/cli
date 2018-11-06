package connect

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/connect"
	"github.com/confluentinc/cli/command/common"
)

// Client handles communication with the service API
var Client connect.Connect

type command struct {
	*cobra.Command
	config  *shared.Config
}

// New returns the Cobra command for Connect.
func New(config *shared.Config, run func(interface{})(error)) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "connect",
			Short: "Manage connect.",
		},
		config: config,
	}
	err := cmd.init(run)
	return cmd.Command, err
}

func (c *command) init(run func(interface{})(error)) error {
	// All commands require login first
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := c.config.CheckLogin(); err != nil {
			return common.HandleError(err, cmd)
		}
		// Lazy load plugin to avoid unnecessarily spawning child processes
		return run(&Client)
	}

	sinkCmd, err := NewSink(c.config)
	if err != nil {
		return err
	}
	c.AddCommand(sinkCmd)

	return nil
}
