package flink

import (
	"time"

	"github.com/spf13/cobra"
)

type statementOut struct {
	CreationDate time.Time `human:"Creation Date" serialized:"creation_date"`
	Name         string    `human:"Name" serialized:"name"`
	Statement    string    `human:"Statement" serialized:"statement"`
	ComputePool  string    `human:"Compute Pool" serialized:"compute_pool"`
	Status       string    `human:"Status" serialized:"status"`
	StatusDetail string    `human:"Status Detail,omitempty" serialized:"status_detail,omitempty"`
}

func (c *command) newStatementCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "statement",
		Short: "Manage Flink SQL statements.",
	}

	cmd.AddCommand(c.newStatementCreateCommand())
	cmd.AddCommand(c.newStatementDeleteCommand())
	cmd.AddCommand(c.newStatementDescribeCommand())
	cmd.AddCommand(c.newStatementExceptionCommand())
	cmd.AddCommand(c.newStatementListCommand())
	cmd.AddCommand(c.newStatementStopCommand())

	return cmd
}

func (c *command) validStatementArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validStatementArgsMultiple(cmd, args)
}

func (c *command) validStatementArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return nil
	}

	listStatementsResponse, err := client.ListStatements(environmentId, c.Context.GetCurrentOrganization(), "", "")
	if err != nil {
		return nil
	}
	statements := listStatementsResponse.GetData()

	suggestions := make([]string, len(statements))
	for i, statement := range statements {
		suggestions[i] = statement.GetName()
	}
	return suggestions
}
