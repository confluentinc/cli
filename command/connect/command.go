package connect

import (
	"os"

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
func New(config *shared.Config, connect Connect) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "connect",
			Short: "Manage connect.",
		},
		config: config,
		connect: connect,
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

	sinkCmd, err := NewSink(c.config, c.connect)
	if err != nil {
		return err
	}
	c.AddCommand(sinkCmd)

	return nil
}
