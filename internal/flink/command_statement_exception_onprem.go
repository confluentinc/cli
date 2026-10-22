package flink

import (
	"github.com/spf13/cobra"
)

type exceptionOutOnPrem struct {
	Timestamp string `human:"Timestamp" serialized:"timestamp"`
	Name      string `human:"Name" serialized:"name"`
	Message   string `human:"Message" serialized:"message"`
}

// No RunRequirement annotation: this group is only ever attached under the on-prem `statement`
// command, whose RequireCloudLogout already covers everything beneath it.
func (c *command) newStatementExceptionCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exception",
		Short: "Manage Flink SQL statement exceptions in Confluent Platform.",
	}

	cmd.AddCommand(c.newStatementExceptionListCommandOnPrem())

	return cmd
}
