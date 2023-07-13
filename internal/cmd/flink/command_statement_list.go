package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newStatementListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink SQL statements.",
		RunE:  c.statementList,
	}
	pcmd.AddCloudFlag(cmd)
	c.addRegionFlag(cmd)
	c.addComputePoolFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) statementList(cmd *cobra.Command, args []string) error {
	client, err := c.GetFlinkGatewayClient()
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	computePoolId, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}

	statements, err := client.ListAllStatements(environmentId, c.Context.LastOrgId, computePoolId)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, statement := range statements {
		list.Add(&statementOut{
			CreationDate: statement.Metadata.GetCreatedAt(),
			Name:         statement.Spec.GetStatementName(),
			Statement:    statement.Spec.GetStatement(),
			ComputePool:  statement.Spec.GetComputePoolId(),
			Status:       statement.Status.GetPhase(),
			StatusDetail: statement.Status.GetDetail(),
		})
	}
	return list.Print()
}
