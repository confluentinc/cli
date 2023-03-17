package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *command) newRegionCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "region",
		Short: "List Flink regions.",
	}

	cmd.AddCommand(c.newRegionListCommand())

	return cmd
}
