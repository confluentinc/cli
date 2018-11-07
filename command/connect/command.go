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

// New returns the default command object for interacting with Connect.
func New(config *shared.Config) (*cobra.Command, error) {
	return newCMD(config, grpcLoader)
}

// NewConnectCommand returns a command object using a custom Connect provider.
func NewConnectCommand(config *shared.Config, provider func(interface{}) error) (*cobra.Command, error) {
	return newCMD(config, provider)
}

// newCMD returns a command for interacting with Connect.
func newCMD(config *shared.Config, provider func(interface{})(error)) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "connect",
			Short: "Manage connect.",
		},
		config: config,
	}
	err := cmd.init(provider)
	return cmd.Command, err
}

// grpcLoader is the default Connect impl provider
func grpcLoader(i interface{}) error {
	return common.LoadPlugin(connect.Name, i)
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
