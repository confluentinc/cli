package flink

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newStatementDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <name-1> [name-2] ... [name-n]",
		Short:             "Delete one or more Flink SQL statements.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgsMultiple),
		RunE:              c.statementDelete,
	}

	pcmd.AddForceFlag(cmd)
	c.addComputePoolFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) statementDelete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClient()
	if err != nil {
		return err
	}

	if confirm, err := c.confirmDeletionStatement(cmd, client, environmentId, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		return client.DeleteStatement(environmentId, id, c.Context.LastOrgId)
	}

	_, err = resource.Delete(args, deleteFunc, resource.FlinkStatement)
	return err
}

func (c *command) confirmDeletionStatement(cmd *cobra.Command, client *ccloudv2.FlinkGatewayClient, environmentId string, args []string) (bool, error) {
	describeFunc := func(id string) error {
		_, err := client.GetStatement(environmentId, id, c.Context.LastOrgId)
		return err
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.FlinkStatement, describeFunc); err != nil {
		return false, err
	}

	return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.FlinkStatement, args))
}
