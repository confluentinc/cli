package flink

import (
	"github.com/spf13/cobra"
)

type flinkEnvironmentOut struct {
	Name        string `human:"Name" serialized:"name"`
	CreatedTime string `human:"Created Time" serialized:"created_time"`
	UpdatedTime string `human:"Updated Time" serialized:"updated_time"`
}

func (c *command) newEnvironmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environment",
		Short:   "Manage Flink Environments",
		Aliases: []string{"env"},
	}
	cmd.AddCommand(c.newEnvironmentListCommand())
	cmd.AddCommand(c.newEnvironmentDeleteommand())
	return cmd
}
