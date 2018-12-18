package connect

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

// New returns the Cobra command for Connect.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Manage connect.",
	}

	cmdFactories := []common.CommandFactory{
		func(config *shared.Config, plugin interface{}) *cobra.Command {
			return NewSink(config, plugin.(Connect))
		},
	}

	err := common.InitCommand("confluent-connect-plugin",
		"connect",
		config,
		cmd,
		cmdFactories)

	return cmd, err
}
