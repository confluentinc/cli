package ksql

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

// New returns the Cobra command for Kafka.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "ksql",
		Short: "Manage ksql.",
	}

	cmdFactories := []common.CommandFactory{
		func(config *shared.Config, plugin interface{}) *cobra.Command {
			return NewClusterCommand(config, plugin.(Ksql))
		},
	}

	err := common.InitCommand("confluent-ksql-plugin",
		"ksql",
		config,
		cmd,
		cmdFactories)

	return cmd, err

}
