package flink

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/properties"
	"github.com/confluentinc/cli/v4/pkg/wait"
)

// Phase enum from flink-gateway/v1 SqlV1StatementStatus.Phase. PENDING /
// FAILING / STOPPING / DELETING are transitioning states; FAILED is the only
// terminal failure; RUNNING / COMPLETED / STOPPED / DEGRADED are terminal
// success. The generator will source the same sets from AsyncConfig.
var (
	flinkStatementPendingPhases = []string{"PENDING", "FAILING", "STOPPING", "DELETING"}
	flinkStatementFailedPhases  = []string{"FAILED"}
)

// flinkStatementCreateWaitTimeout is the default for --wait-timeout. It is kept
// at 1 minute to match the pre-framework `--wait` behavior (a hardcoded
// retry.Retry(time.Second, time.Minute, ...) poll), so customers who scripted
// around that 1-minute timer are not affected by the --wait refactor. Bump to
// 6h in the next major version (aligning with terraform-provider-confluent's
// statementsAPICreateTimeout in internal/provider/constants.go); CLI users can
// already lengthen it with --wait-timeout in the meantime. (APIE-1040)
const flinkStatementCreateWaitTimeout = time.Minute

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
	cmd.Flags().Bool("wait", false, "Block until the statement reaches a terminal state.")
	cmd.Flags().Duration("wait-timeout", flinkStatementCreateWaitTimeout, "Maximum time to wait when --wait is set.")
	cmd.Flags().StringSlice("property", []string{}, "A mechanism to pass properties in the form key=value when creating a Flink statement.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)

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
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	if computePool == "" {
		if cloud == "" || region == "" {
			return errors.New("Flink cloud and region flags are required when compute pool is not specified.")
		}
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
			Statement:  flinkgatewayv1.PtrString(sql),
			Properties: &statementProperties,
		},
	}
	var client *ccloudv2.FlinkGatewayClient
	if computePool != "" {
		statement.Spec.ComputePoolId = flinkgatewayv1.PtrString(computePool)
		client, err = c.GetFlinkGatewayClient(true)
		if err != nil {
			return err
		}
	} else {
		client, err = c.GetFlinkGatewayClient(false)
		if err != nil {
			return err
		}
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
	shouldWait, err := cmd.Flags().GetBool("wait")
	if err != nil {
		return err
	}
	if shouldWait {
		timeout, err := cmd.Flags().GetDuration("wait-timeout")
		if err != nil {
			return err
		}
		statement, err = wait.PollPhases(cmd.Context(), wait.PhaseOptions[flinkgatewayv1.SqlV1Statement]{
			Fetch: func() (flinkgatewayv1.SqlV1Statement, error) {
				return client.GetStatement(environmentId, name, c.Context.LastOrgId)
			},
			Phase:         func(s flinkgatewayv1.SqlV1Statement) string { return s.Status.GetPhase() },
			PendingPhases: flinkStatementPendingPhases,
			FailedPhases:  flinkStatementFailedPhases,
			PollInterval:  time.Second,
			Timeout:       timeout,
		})
		if err != nil {
			// wait.ErrFailed and wait.ErrTimeout are package-level sentinels
			// returned directly from Poll (never wrapped); a direct == compare
			// is correct and avoids importing stdlib errors alongside the CLI
			// errors package.
			switch err {
			case wait.ErrFailed:
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(`statement "%s" entered failed phase %q: %s`, name, statement.Status.GetPhase(), statement.Status.GetDetail()),
					fmt.Sprintf("Inspect the statement with `confluent flink statement describe %s`.", name),
				)
			case wait.ErrTimeout:
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(`wait timed out: statement "%s" is still in phase %q`, name, statement.Status.GetPhase()),
					"Increase `--wait-timeout` or omit `--wait`.",
				)
			default:
				// Fetch error or context cancellation — surface the underlying
				// cause unmodified; suggesting a longer timeout would mislead.
				return err
			}
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
