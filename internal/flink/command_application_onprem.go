package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

type flinkApplicationSummary struct {
	Name        string `human:"Name" serialized:"name"`
	Environment string `human:"Environment" serialized:"environment"`
	JobId       string `human:"Job ID" serialized:"job_id"`
	JobState    string `human:"Job State" serialized:"job_state"`
}

func (c *command) newApplicationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "application",
		Short:       "Manage Flink application",
		Aliases:     []string{"app"},
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newApplicationCreateCommand())
	cmd.AddCommand(c.newApplicationDeleteCommand())
	cmd.AddCommand(c.newApplicationDescribeCommand())
	cmd.AddCommand(c.newApplicationListCommand())
	cmd.AddCommand(c.newApplicationUpdateCommand())
	cmd.AddCommand(c.newApplicationWebUiForwardCommand())

	return cmd
}
