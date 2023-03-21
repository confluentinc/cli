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

func (c *command) newStatementDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a Flink SQL statement.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.statementCreate,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) statementDelete(cmd *cobra.Command, args []string) error {
	statement, err := c.V2Client.GetStatement(c.EnvironmentId(), args[0])
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmYesNoMsg, resource.FlinkStatement, args[0])
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, statement.Spec.GetStatementName()); err != nil || !ok {
		return err
	}

	if err := c.V2Client.DeleteStatement(c.EnvironmentId(), args[0]); err != nil {
		return err
	}

	output.Printf(errors.DeletedResourceMsg, resource.FlinkStatement, args[0])
	return nil
}
