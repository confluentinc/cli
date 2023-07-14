package flink

import (
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/flink/config"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newStatementCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Flink SQL statement.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.statementCreate,
	}

	c.addDatabaseFlag(cmd)
	c.addComputePoolFlag(cmd)
	cmd.Flags().String("identity-pool", "", "Identity pool ID.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) statementCreate(cmd *cobra.Command, args []string) error {
	client, err := c.GetFlinkGatewayClient()
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	properties := map[string]string{}

	database, err := cmd.Flags().GetString("database")
	if err != nil {
		return err
	}
	if database != "" {
		properties[config.ConfigKeyDatabase] = database
	}

	computePoolId := c.Context.GetCurrentFlinkComputePool()
	if computePoolId == "" {
		return errors.NewErrorWithSuggestions("no compute pool selected", "Select a compute pool with `confluent flink compute-pool use` or `--compute-pool`.")
	}

	identityPoolId := c.Context.GetCurrentIdentityPool()
	if identityPoolId == "" {
		return errors.NewErrorWithSuggestions("no identity pool selected", "Select an identity pool with `confluent iam pool use` or `--identity-pool`.")
	}

	statement := flinkgatewayv1alpha1.SqlV1alpha1Statement{Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
		StatementName:  flinkgatewayv1alpha1.PtrString(uuid.New().String()[:18]),
		Statement:      flinkgatewayv1alpha1.PtrString(args[0]),
		Properties:     &properties,
		ComputePoolId:  flinkgatewayv1alpha1.PtrString(computePoolId),
		IdentityPoolId: flinkgatewayv1alpha1.PtrString(identityPoolId),
	}}

	statement, err = client.CreateStatement(environmentId, statement, c.Context.LastOrgId)
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
