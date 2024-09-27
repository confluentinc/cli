package flink

import (
	"github.com/spf13/cobra"
)

type flinkEnvironmentSummary struct {
	Name        string `human:"Name" serialized:"name"`
	CreatedTime string `human:"Created Time" serialized:"created_time"`
	UpdatedTime string `human:"Updated Time" serialized:"updated_time"`
}

type flinkEnvironmentOutput struct {
	Name        string `human:"Name" serialized:"name"`
	CreatedTime string `human:"Created Time" serialized:"created_time"`
	UpdatedTime string `human:"Updated Time" serialized:"updated_time"`
	Defaults    string `human:"Defaults" serialized:"defaults"`
}

func (c *unauthenticatedCommand) newEnvironmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environment",
		Short:   "Manage Flink Environments",
		Aliases: []string{"env"},
	}

	cmd.AddCommand(c.newEnvironmentCreateCommand())
	cmd.AddCommand(c.newEnvironmentDeleteCommand())
	cmd.AddCommand(c.newEnvironmentListCommand())
	cmd.AddCommand(c.newEnvironmentUpdateCommand())
	return cmd
}
