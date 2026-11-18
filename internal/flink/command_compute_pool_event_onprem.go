package flink

import (
	"github.com/spf13/cobra"
)

// No RunRequirement annotation: this group is only ever attached under the on-prem `compute-pool` command, whose
// RequireCloudLogout already covers everything beneath it.
func (c *command) newComputePoolEventCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "Manage Flink compute pool events in Confluent Platform.",
	}

	cmd.AddCommand(c.newComputePoolEventListCommandOnPrem())

	return cmd
}
