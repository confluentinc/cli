package flink

import (
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/flink"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newStatementCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a Flink SQL statement.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.statementCreate,
	}

	cmd.Flags().String("sql", "", "The Flink SQL statement.")
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddCloudFlag(cmd)
	c.addRegionFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("sql"))

	return cmd
}

func (c *command) statementCreate(cmd *cobra.Command, args []string) error {
	serviceAccount := c.Context.GetCurrentServiceAccount()
	if serviceAccount == "" {
		output.ErrPrintln(flink.ServiceAccountWarning)
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	name := uuid.New().String()
	if len(args) == 1 {
		name = args[0]
	}

	sql, err := cmd.Flags().GetString("sql")
	if err != nil {
		return err
	}

	statement := flinkgatewayv1alpha1.SqlV1alpha1Statement{Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
		StatementName: flinkgatewayv1alpha1.PtrString(name),
		Statement:     flinkgatewayv1alpha1.PtrString(sql),
	}}

	client, err := c.GetFlinkGatewayClient()
	if err != nil {
		return err
	}

	statement, err = client.CreateStatement(serviceAccount, environmentId, statement, c.Context.LastOrgId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&statementOut{
		CreationDate: statement.Metadata.GetCreatedAt(),
		Name:         statement.Spec.GetStatementName(),
		Statement:    statement.Spec.GetStatement(),
		ComputePool:  statement.Spec.GetComputePoolId(),
		Status:       statement.Status.GetPhase(),
		StatusDetail: statement.Status.GetDetail(),
	})
	return table.Print()
}
