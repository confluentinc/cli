package flink

import (
	"github.com/spf13/cobra"
)

type flinkEnvironmentOut struct {
	Name            string `human:"Name" serialized:"name"`
	DefaultStrategy string `human:"Default Strategy" serialized:"default_strategy"`
	CreatedTime     string `human:"Created Time" serialized:"created_time"`
	UpdatedTime     string `human:"Updated Time" serialized:"updated_time"`
}

func (c *command) newEnvironmentOnPremCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environment",
		Short:   "Manage Flink Environments",
		Aliases: []string{"env"},
	}
	cmd.AddCommand(c.newEnvironmentUseCommand())
	cmd.AddCommand(c.newEnvironmentListOnPremCommand())
	cmd.AddCommand(c.newEnvironmentCreateCommandOnPrem())
	cmd.AddCommand(c.newEnvironmentUpdateCommandOnPrem())
	cmd.AddCommand(c.newEnvironmentDescribeCommandOnPrem())
	cmd.AddCommand(c.newEnvironmentDeleteOnPremCommand())
	return cmd
}
