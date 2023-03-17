package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newSqlStatementDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a Flink SQL statement.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.sqlStatementCreate,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) sqlStatementDelete(cmd *cobra.Command, args []string) error {
	sqlStatement, err := c.V2Client.GetSqlStatement(c.EnvironmentId(), args[0])
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmYesNoMsg, resource.FlinkSqlStatement, args[0])
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, sqlStatement.Spec.GetStatementName()); err != nil || !ok {
		return err
	}

	if err := c.V2Client.DeleteSqlStatement(c.EnvironmentId(), args[0]); err != nil {
		return err
	}

	output.Printf(errors.DeletedResourceMsg, resource.FlinkSqlStatement, args[0])
	return nil
}
