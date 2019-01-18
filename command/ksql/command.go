package ksql

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/ksql"
)

type command struct {
	*cobra.Command
	config *shared.Config
}

// New returns the default command object for interacting with KSQL.
func New(config *shared.Config) (*cobra.Command, error) {
	return newCMD(config, common.GRPCLoader(ksql.Name))
}

// NewKSQLCommand returns a command object using a custom KSQL provider.
func NewKSQLCommand(config *shared.Config, provider func(interface{}) error) (*cobra.Command, error) {
	return newCMD(config, provider)
}

// newCMD returns a command for interacting with KSQL.
func newCMD(config *shared.Config, run func(interface{}) error) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "ksql",
			Short: "Manage KSQL.",
		},
		config: config,
	}
	err := cmd.init(run)
	return cmd.Command, err
}

func (c *command) init(plugin common.Provider) error {
	c.AddCommand(NewClusterCommand(c.config, plugin))
	return nil
}
