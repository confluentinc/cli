package ksql

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/ksql"
)

// Client handles communication with the service API
var Client ksql.Ksql

type command struct {
	*cobra.Command
	config *shared.Config
}

// New returns the default command object for interacting with KSQL.
func New(config *shared.Config) (*cobra.Command, error) {
	return newCMD(config, grpcLoader)
}

// NewKSQLCommand returns a command object using a custom KSQL provider.
func NewKSQLCommand(config *shared.Config, provider func(interface{}) error) (*cobra.Command, error) {
	return newCMD(config, provider)
}

// New returns a command for interacting with KSQL.
func newCMD(config *shared.Config, run func(interface{})(error)) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "ksql",
			Short: "Manage ksql.",
		},
		config: config,
	}
	err := cmd.init(run)
	return cmd.Command, err
}

// grpcLoader is the default KSQL impl provider
func grpcLoader(i interface{}) error {
	return common.DefaultClient(ksql.Name)(i)
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

	c.AddCommand(NewClusterCommand(c.config))
	return nil
}
