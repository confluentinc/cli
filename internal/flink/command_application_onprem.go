package flink

import (
	"github.com/spf13/cobra"
)

type flinkApplicationOut struct {
	Name        string `human:"Name" serialized:"name"`
	Environment string `human:"Environment" serialized:"environment"`
	JobId       string `human:"Job ID" serialized:"job_id"`
	JobState    string `human:"Job State" serialized:"job_state"`
}

func (c *command) newApplicationCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application",
		Short:   "Manage Flink Application",
		Aliases: []string{"app"},
	}

	cmd.AddCommand(c.newApplicationListCommandOnPrem())
	return cmd
}
