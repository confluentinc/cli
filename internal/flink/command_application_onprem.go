package flink

import (
	"github.com/spf13/cobra"
)

type flinkApplicationSummary struct {
	Name        string `human:"Name" serialized:"name"`
	Environment string `human:"Environment" serialized:"environment"`
	JobId       string `human:"Job ID" serialized:"job_id"`
	JobState    string `human:"Job State" serialized:"job_state"`
}

type flinkApplicationOutput struct {
	ApiVersion string `human:"API Version" serialized:"api_version"`
	Kind       string `human:"Kind" serialized:"kind"`
	Metadata   string `human:"Metadata" serialized:"metadata"`
	Spec       string `human:"Spec" serialized:"spec"`
	Status     string `human:"Status" serialized:"status"`
}

func (c *unauthenticatedCommand) newApplicationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application",
		Short:   "Manage Flink Application",
		Aliases: []string{"app"},
	}

	cmd.AddCommand(c.newApplicationCreateCommand())
	cmd.AddCommand(c.newApplicationDescribeCommand())
	cmd.AddCommand(c.newApplicationDeleteCommand())
	cmd.AddCommand(c.newApplicationListCommand())
	cmd.AddCommand(c.newApplicationUpdateCommand())

	return cmd
}
