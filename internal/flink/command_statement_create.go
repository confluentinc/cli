package flink

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/properties"
	"github.com/confluentinc/cli/v4/pkg/retry"
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
				Text: `Create a Flink SQL statement named "my-statement" in compute pool "lfcp-123456" with service account "sa-123456", using Kafka cluster "my-cluster" as the default database, and with additional properties.`,
				Code: `confluent flink statement create my-statement --sql "SELECT * FROM my-topic;" --compute-pool lfcp-123456 --service-account sa-123456 --database my-cluster --property property1=value1,property2=value2`,
			},
		),
	}

	cmd.Flags().String("sql", "", "The Flink SQL statement.")
	c.addComputePoolFlag(cmd)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	c.addDatabaseFlag(cmd)
	cmd.Flags().Bool("wait", false, "Block until the statement is running or has failed.")
	cmd.Flags().StringSlice("property", []string{}, "A mechanism to pass properties in the form key=value when creating a Flink statement.")
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

	statementProperties := map[string]string{config.KeyCatalog: environment.GetDisplayName()}
	if database != "" {
		statementProperties[config.KeyDatabase] = database
	}

	// Parse custom properties if provided
	configs, err := cmd.Flags().GetStringSlice("property")
	if err != nil {
		return err
	}

	if len(configs) > 0 {
		configMap, err := properties.ConfigSliceToMap(configs)
		if err != nil {
			return err
		}
		for key, value := range configMap {
			statementProperties[key] = value
		}
	}

	statement := flinkgatewayv1.SqlV1Statement{
		Name: flinkgatewayv1.PtrString(name),
		Spec: &flinkgatewayv1.SqlV1StatementSpec{
			Statement:     flinkgatewayv1.PtrString(sql),
			Properties:    &statementProperties,
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

		// If the statement produces results, fetch and display them
		traits := statement.Status.GetTraits()
		schema := traits.GetSchema()
		if columns := schema.GetColumns(); len(columns) > 0 {
			statementResults, err := fetchAllResults(client, environmentId, name, c.Context.LastOrgId, schema, 0)
			if err != nil {
				return err
			}
			return printStatementResults(cmd, statementResults)
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
	table.Filter([]string{"CreationDate", "Name", "Statement", "ComputePool", "Status", "StatusDetail"})
	return table.Print()
}
