package flink

import (
	"github.com/spf13/cobra"
)

type flinkApplicationSummaryOut struct {
	Name        string `human:"Name" serialized:"name"`
	Environment string `human:"Environment" serialized:"environment"`
	JobName     string `human:"Job Name" serialized:"job_name"`
	JobStatus   string `human:"Job Status" serialized:"job_status"`
}

func (c *command) newApplicationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application",
		Short:   "Manage Flink applications.",
		Aliases: []string{"app"},
	}

	cmd.AddCommand(c.newApplicationCreateCommand())
	cmd.AddCommand(c.newApplicationDeleteCommand())
	cmd.AddCommand(c.newApplicationDescribeCommand())
	cmd.AddCommand(c.newApplicationListCommand())
	cmd.AddCommand(c.newApplicationUpdateCommand())
	cmd.AddCommand(c.newApplicationWebUiForwardCommand())

	return cmd
}
