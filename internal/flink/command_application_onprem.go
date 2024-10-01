package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

type flinkApplicationOut struct {
	Name        string `human:"Name" serialized:"name"`
	Environment string `human:"Environment" serialized:"environment"`
	JobId       string `human:"Job ID" serialized:"job_id"`
	JobState    string `human:"Job State" serialized:"job_state"`
}

func (c *command) newApplicationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "application",
		Short:       "Manage Flink Application",
		Aliases:     []string{"app"},
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newApplicationListCommand())
	cmd.AddCommand(c.newApplicationDeleteCommand())

	return cmd
}
