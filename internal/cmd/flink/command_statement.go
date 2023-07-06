package flink

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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

	cmd.AddCommand(c.newStatementDeleteCommand())
	cmd.AddCommand(c.newStatementListCommand())
	cmd.AddCommand(c.newStatementGetDetailsCommand())
	cmd.AddCommand(c.newExceptionsCommand())

	return cmd
}

func (c *command) addComputePoolFlag(cmd *cobra.Command) {
	cmd.Flags().String("compute-pool", "", "Flink compute pool ID.")

	pcmd.RegisterFlagCompletionFunc(cmd, "compute-pool", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return c.autocompleteComputePools()
	})
}

func (c *command) validStatementArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	client, err := c.GetFlinkGatewayClient()
	if err != nil {
		return nil
	}

	listStatementsResponse, err := client.ListStatements(environmentId, c.Context.LastOrgId, "")
	if err != nil {
		return nil
	}
	statements := listStatementsResponse.GetData()

	suggestions := make([]string, len(statements))
	for i, statement := range statements {
		suggestions[i] = fmt.Sprintf("%s\t%s", statement.Spec.GetStatementName(), statement.Spec.GetStatement())
	}
	return suggestions
}
