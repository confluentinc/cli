package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newStatementDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <name-1> [name-2] ... [name-n]",
		Short:             "Delete one or more Flink SQL statements.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgsMultiple),
		RunE:              c.statementDelete,
	}

	pcmd.AddCloudFlag(cmd)
	c.addRegionFlag(cmd)
	pcmd.AddForceFlag(cmd)
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

	existenceFunc := func(id string) bool {
		_, err := client.GetStatement(environmentId, id, c.Context.LastOrgId)
		return err == nil
	}

	if err := deletion.ValidateAndConfirmDeletionYesNo(cmd, args, existenceFunc, resource.FlinkStatement); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return client.DeleteStatement(environmentId, id, c.Context.LastOrgId)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.FlinkStatement)
	return err
}
