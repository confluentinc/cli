package connect

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/connect"
	"github.com/confluentinc/cli/command/common"
)

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

func (c *command) init(plugin common.Provider) error {
	sinkCmd, err := NewSink(c.config, plugin)
	if err != nil {
		return err
	}
	c.AddCommand(sinkCmd)

	return nil
}
