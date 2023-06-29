package flink

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newStatementListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink SQL statements.",
		RunE:  c.statementList,
	}

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

	statements, err := client.ListAllStatements(environmentId, c.Context.LastOrgId)
	if err != nil {
		return err
	}

	sortStatementsByCreationTime(statements)

	list := output.NewList(cmd)
	list.Sort(false) // disable the default sort
	for _, statement := range statements {
		list.Add(&statementOut{
			Name:         statement.Spec.GetStatementName(),
			Statement:    statement.Spec.GetStatement(),
			ComputePool:  statement.Spec.GetComputePoolId(),
			Status:       statement.Status.GetPhase(),
			StatusDetail: statement.Status.GetDetail(),
		})
	}
	return list.Print()
}

func sortStatementsByCreationTime(statements []v1alpha1.SqlV1alpha1Statement) {
	// sort ascending by creation time - with most recent statement being listed at the bottom of the table
	sort.Slice(statements, func(i, j int) bool {
		metadataStatement1 := statements[i].GetMetadata()
		metadataStatement2 := statements[j].GetMetadata()
		return metadataStatement1.GetCreatedAt().Before(metadataStatement2.GetCreatedAt())
	})
}
