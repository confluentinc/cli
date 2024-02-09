package flink

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/flink/config"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/retry"
)

func (c *command) newStatementCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a Flink SQL statement.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.statementCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a Flink SQL statement in the current compute pool.",
				Code: `confluent flink statement create --sql "SELECT * FROM table;"`,
			},
			examples.Example{
				Text: `Create a Flink SQL statement named "my-statement" in compute pool "lfcp-123456" with service account "sa-123456" and using Kafka cluster "my-cluster" as the default database.`,
				Code: `confluent flink statement create my-statement --sql "SELECT * FROM my-topic;" --compute-pool lfcp-123456 --service-account sa-123456 --database my-cluster`,
			},
		),
	}

	cmd.Flags().String("sql", "", "The Flink SQL statement.")
	c.addComputePoolFlag(cmd)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	c.addDatabaseFlag(cmd)
	cmd.Flags().Bool("wait", false, "Block until the statement is running or has failed.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("sql"))

	return cmd
}

func (c *command) statementCreate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	environment, err := c.V2Client.GetOrgEnvironment(environmentId)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), "List available environments with `confluent environment list`.")
	}

	computePool := c.Context.GetCurrentFlinkComputePool()
	if computePool == "" {
		return errors.NewErrorWithSuggestions(
			"no compute pool selected",
			"Select a compute pool with `confluent flink compute-pool use` or `--compute-pool`.",
		)
	}

	name := types.GenerateStatementName()
	if len(args) == 1 {
		name = args[0]
	}

	sql, err := cmd.Flags().GetString("sql")
	if err != nil {
		return err
	}

	database, err := cmd.Flags().GetString("database")
	if err != nil {
		return err
	}

	properties := map[string]string{config.KeyCatalog: environment.GetDisplayName()}
	if database != "" {
		properties[config.KeyDatabase] = database
	}

	statement := flinkgatewayv1.SqlV1Statement{
		Name: flinkgatewayv1.PtrString(name),
		Spec: &flinkgatewayv1.SqlV1StatementSpec{
			Statement:     flinkgatewayv1.PtrString(sql),
			Properties:    &properties,
			ComputePoolId: flinkgatewayv1.PtrString(computePool),
		},
	}

	client, err := c.GetFlinkGatewayClient(true)
	if err != nil {
		return err
	}

	serviceAccount, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}

	principal := serviceAccount
	if serviceAccount == "" {
		output.ErrPrintln(c.Config.EnableColor, serviceAccountWarning)
		principal = c.Context.GetUser().GetResourceId()
	}

	statement, err = client.CreateStatement(statement, principal, environmentId, c.Context.LastOrgId)
	if err != nil {
		return err
	}

	wait, err := cmd.Flags().GetBool("wait")
	if err != nil {
		return err
	}
	if wait {
		err := retry.Retry(time.Second, time.Minute, func() error {
			statement, err = client.GetStatement(environmentId, name, c.Context.LastOrgId)
			if err != nil {
				return err
			}

			if statement.Status.GetPhase() == "PENDING" {
				return fmt.Errorf(`statement phase is "%s"`, statement.Status.GetPhase())
			}

			return nil
		})
		if err != nil {
			return err
		}
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
