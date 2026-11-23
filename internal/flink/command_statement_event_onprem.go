package flink

import (
	"github.com/spf13/cobra"
)

// No RunRequirement annotation: this group is only ever attached under the on-prem `statement` command, whose
// RequireCloudLogout already covers everything beneath it.
func (c *command) newStatementEventCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "Manage Flink SQL statement events in Confluent Platform.",
	}

	cmd.AddCommand(c.newStatementEventListCommandOnPrem())

	return cmd
}
