package flink

import (
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	flinkgatewayv1beta1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1beta1"

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
	c.addComputePoolFlag(cmd)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
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

	statement := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: flinkgatewayv1beta1.PtrString(name),
		Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{Statement: flinkgatewayv1beta1.PtrString(sql)},
	}

	client, err := c.GetFlinkGatewayClient(true)
	if err != nil {
		return err
	}

	statement, err = client.CreateStatement(statement, serviceAccount, environmentId, c.Context.LastOrgId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&statementOut{
		CreationDate: statement.Metadata.GetCreatedAt(),
		Name:         statement.GetName(),
		Statement:    statement.Spec.GetStatement(),
		ComputePool:  statement.Spec.GetComputePoolId(),
		Status:       statement.Status.GetPhase(),
		StatusDetail: statement.Status.GetDetail(),
	})
	return table.Print()
}
