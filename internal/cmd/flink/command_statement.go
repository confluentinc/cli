package flink

import "github.com/spf13/cobra"

type statementOut struct {
	Name         string `human:"Name" serialized:"name"`
	Statement    string `human:"Statement" serialized:"statement"`
	ComputePool  string `human:"Compute Pool" serialized:"compute_pool"`
	Status       string `human:"Status" serialized:"status"`
	StatusDetail string `human:"Status Detail,omitempty" serialized:"status_detail,omitempty"`
}

func (c *command) newStatementCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "statement",
		Short: "Manage Flink SQL statements.",
	}

	cmd.AddCommand(c.newStatementDeleteCommand())
	cmd.AddCommand(c.newStatementListCommand())

	return cmd
}
