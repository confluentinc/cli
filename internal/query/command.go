package query

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	cliconfig "github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/query"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/jwt"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/properties"
	"github.com/confluentinc/cli/v4/pkg/wait"
)

const (
	// snapshotModeProperty makes this a bounded, point-in-time read; not overridable via --property.
	snapshotModeProperty = "sql.snapshot.mode"
	snapshotModeNow      = "now"

	// queryFeatureFlag gates the command's visibility.
	queryFeatureFlag = "cli.query"
)

type command struct {
	*pcmd.AuthenticatedCLICommand

	// authTokenMu guards client.AuthToken against a leaked refresh racing a stop attempt.
	authTokenMu sync.Mutex
}

// New mounts `confluent flink query`: the entry point for a bounded, one-shot Flink
// SQL read, with no engine routing to other backends.
func New(cfg *cliconfig.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Run a bounded Flink SQL query and print its results.",
		Long: "Run a bounded (snapshot) Flink SQL query, block until it finishes, and print the complete result set.\n\n" +
			"The SQL is given with `--sql`, or with `--file` to read it from a file.\n\n" +
			"Unlike statement creation, which submits a statement and returns immediately, this command waits for every " +
			"result page and exits with a non-zero status if the statement fails. It is intended for scripting and " +
			"one-shot queries against a bounded (point-in-time) result set.\n\n" +
			"With `-o json` or `-o yaml`, output defaults to an envelope carrying the column schema alongside the rows, " +
			"since the rows on their own carry no type information. Pass `--raw` for a bare array of row objects instead.",
		Args: cobra.NoArgs,
		// Hidden until the flag targets an org; cfg.IsTest keeps it visible to the
		// integration suite regardless of the (unreachable in tests) LD evaluation.
		Hidden: !(cfg.IsTest || featureflags.Manager.BoolVariation(queryFeatureFlag, cfg.Context(), cliconfig.CliLaunchDarklyClient, true, false)),
		Annotations: map[string]string{
			pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin,
		},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Run a bounded query in the current compute pool and print the rows as a table.",
				Code: `confluent flink query --sql "SELECT * FROM orders LIMIT 10;"`,
			},
			examples.Example{
				Text: "Run a bounded query against Kafka cluster \"my-cluster\" and emit JSON for a script to consume.",
				Code: `confluent flink query --sql "SELECT status, COUNT(*) FROM orders GROUP BY status;" --compute-pool lfcp-123456 --database my-cluster --output json`,
			},
			examples.Example{
				Text: "Emit a bare JSON array of rows, with no envelope, for a script that only wants the data.",
				Code: `confluent flink query --sql "SELECT * FROM orders LIMIT 10;" --output json --raw`,
			},
		),
	}

	c := &command{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.runQuery

	cmd.Flags().String("sql", "", `The Flink SQL statement. Alternatively, pass it with "-f".`)
	cmd.Flags().StringP("file", "f", "", `Path to a file containing the Flink SQL statement. Alternatively, pass the SQL with "--sql".`)
	pcmd.AddComputePoolFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddDatabaseFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().StringSlice("property", []string{}, "A mechanism to pass properties in the form key=value when creating a Flink statement.")
	cmd.Flags().Duration("wait-timeout", config.DefaultTimeoutDuration, "Maximum time to wait for the query to finish before giving up.")
	cmd.Flags().Int("max-rows", 0, "Stop fetching and discard the rest after this many rows, or 0 to fetch every row. Client-side only: rows past the limit are still produced by the query.")
	cmd.Flags().Bool("raw", false, `Emit the rows as a bare array with no envelope. Requires "-o json" or "-o yaml".`)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	c.addCatalogAlias(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)

	cmd.MarkFlagsOneRequired("sql", "file")
	cmd.MarkFlagsMutuallyExclusive("sql", "file")
	cmd.MarkFlagsMutuallyExclusive("environment", "catalog")

	return cmd
}

// addCatalogAlias shares --environment's pflag.Value directly, since --environment
// already persists to context and ParseFlagsIntoContext reads it before RunE runs.
func (c *command) addCatalogAlias(cmd *cobra.Command) {
	environmentFlag := cmd.Flags().Lookup("environment")
	cmd.Flags().Var(environmentFlag.Value, "catalog", `Alias for "--environment".`)
}

func (c *command) runQuery(cmd *cobra.Command, _ []string) error {
	// Registered before any other work to shrink the window where a Ctrl-C
	// predates any signal handling and falls through to the OS default
	// disposition (silent kill). Doesn't close it entirely — PersistentPreRunE's
	// own network calls run before this line even executes.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	sql, err := resolveSQL(cmd)
	if err != nil {
		return err
	}

	database, err := c.resolveDatabase(cmd)
	if err != nil {
		return err
	}

	timeout, maxRows, raw, err := parseQueryFlags(cmd)
	if err != nil {
		return err
	}

	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	client, name, err := c.createQueryStatement(ctx, cmd, environmentId, database, sql)
	if err != nil {
		return err
	}

	// The statement now exists server-side; this defer stops it unless settled is
	// set, so no exit path can forget to release the compute it's holding.
	// announceStop stays false only for the routine case: a fully successful,
	// untruncated drain landing on a non-terminal phase (expected for Kafka-backed
	// sources). It's set true below for --max-rows truncation and for any error,
	// since both are cases the user needs to know the cleanup outcome of.
	settled := false
	announceStop := false
	defer func() {
		if settled {
			return
		}
		if announceStop {
			c.stopStatementAndReport(client, environmentId, name)
		} else if ok, err := c.stopStatement(client, environmentId, name); !ok {
			// Even the routine, deliberately-quiet cleanup must surface a *failed*
			// stop: the statement is still running and burning compute, which the
			// user needs to know even though the query itself succeeded. Only a
			// successful quiet stop stays silent (see bug #2).
			reportStopFailure(name, err)
		}
	}()

	options := query.Options{
		Client:         client,
		EnvironmentId:  environmentId,
		OrganizationId: c.Context.LastOrgId,
		MaxRows:        maxRows,
		RequireBounded: true,
		RefreshToken:   c.refreshGatewayToken(client, jwt.NewValidator()),
	}

	result, err := query.Run(ctx, options, name)
	if err != nil {
		// If handleQueryError leaves settled false (any error it doesn't already
		// stop and announce itself), the deferred cleanup above must still speak
		// up: the user just saw an error and needs to know whether the leftover
		// statement was actually cleaned up, not have that outcome logged quietly.
		announceStop = true
		return c.handleQueryError(cmd, client, environmentId, name, err, &settled)
	}

	// drain() refreshes result.Statement, so this reflects reality even after
	// Truncated — a job can stay RUNNING after its last row ships either way.
	settled = query.IsTerminal(result.Phase())
	announceStop = result.Truncated

	// STOPPED/DELETING here means something other than us ended the statement.
	if err := phaseError(name, result); err != nil {
		return err
	}

	if result.Truncated {
		output.ErrPrintf(false, "Warning: stopped after %d rows because of the `--max-rows` flag. The result set below is truncated.\n", maxRows)
	}

	isAppendOnly, appendOnlyKnown := warnIfChangelog(result)
	if err := c.printQueryResult(cmd, name, result, isAppendOnly, appendOnlyKnown, raw); err != nil {
		// A failed print is still an error the user sees, so the deferred cleanup
		// must announce the stop outcome like every other error path — otherwise
		// the user is left an "Error:" with no word on whether the statement,
		// which may still be RUNNING, was released.
		announceStop = true
		return err
	}
	return nil
}

// parseQueryFlags reads and validates the numeric/output flags that gate the run
// before any network call.
func parseQueryFlags(cmd *cobra.Command) (time.Duration, int, bool, error) {
	timeout, err := cmd.Flags().GetDuration("wait-timeout")
	if err != nil {
		return 0, 0, false, err
	}
	if timeout <= 0 {
		return 0, 0, false, errors.New("the `--wait-timeout` flag must be positive")
	}

	maxRows, err := cmd.Flags().GetInt("max-rows")
	if err != nil {
		return 0, 0, false, err
	}
	if maxRows < 0 {
		return 0, 0, false, errors.New("the `--max-rows` flag must not be negative")
	}

	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return 0, 0, false, err
	}
	if raw && !output.GetFormat(cmd).IsSerialized() {
		return 0, 0, false, errors.New("the `--raw` flag requires `-o json` or `-o yaml`")
	}

	return timeout, maxRows, raw, nil
}

// createQueryStatement resolves the environment and gateway client, builds the
// bounded snapshot statement, and submits it, returning the client and the
// generated statement name. A cancellation during any of these pre-result steps
// is mapped to interruptedError.
func (c *command) createQueryStatement(ctx context.Context, cmd *cobra.Command, environmentId, database, sql string) (*ccloudv2.FlinkGatewayClient, string, error) {
	environment, err := wait.Call(ctx, func() (orgv2.OrgV2Environment, error) {
		return c.V2Client.GetOrgEnvironment(environmentId)
	})
	if err != nil {
		return nil, "", c.interruptOr(cmd, err, "", false,
			errors.NewErrorWithSuggestions(err.Error(), "List available environments with `confluent environment list`."))
	}

	statementProperties, err := c.buildQueryProperties(cmd, environment.GetDisplayName(), database)
	if err != nil {
		return nil, "", err
	}

	name := types.GenerateStatementName()
	statement := flinkgatewayv1.SqlV1Statement{
		Name: flinkgatewayv1.PtrString(name),
		Spec: &flinkgatewayv1.SqlV1StatementSpec{
			Statement:  flinkgatewayv1.PtrString(sql),
			Properties: &statementProperties,
		},
	}

	// computePool falls back to context (`flink compute-pool use`);
	// GetFlinkGatewayClient decides whether pool or cloud/region is enough.
	computePool := c.Context.GetCurrentFlinkComputePool()
	if computePool != "" {
		statement.Spec.ComputePoolId = flinkgatewayv1.PtrString(computePool)
	}
	client, err := wait.Call(ctx, func() (*ccloudv2.FlinkGatewayClient, error) {
		return c.GetFlinkGatewayClient(computePool != "")
	})
	if err != nil {
		// No statement exists yet at this point, same as the environment lookup above.
		return nil, "", c.interruptOr(cmd, err, "", false, err)
	}

	principal, err := c.resolvePrincipal(cmd)
	if err != nil {
		return nil, "", err
	}

	stopped, err := createStatement(ctx, createStatementGracePeriod, func() (flinkgatewayv1.SqlV1Statement, error) {
		return client.CreateStatement(statement, principal, environmentId, c.Context.LastOrgId)
	}, func() bool {
		stopped, _ := c.stopStatement(client, environmentId, name)
		return stopped
	})
	if err != nil {
		return nil, "", c.interruptOr(cmd, err, name, stopped, err)
	}

	return client, name, nil
}

// resolvePrincipal is the statement's spec.principal: the --service-account when
// given, otherwise the logged-in user.
func (c *command) resolvePrincipal(cmd *cobra.Command) (string, error) {
	serviceAccount, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return "", err
	}
	if serviceAccount != "" {
		return serviceAccount, nil
	}
	return c.Context.GetUser().GetResourceId(), nil
}

// interruptOr maps a pre-result error to interruptedError when it's a Ctrl-C/
// timeout, and to fallback otherwise. name is "" (and stopped ignored) before a
// statement exists.
func (c *command) interruptOr(cmd *cobra.Command, err error, name string, stopped bool, fallback error) error {
	if isInterrupted(err) {
		return interruptedError(cmd, err, name, stopped)
	}
	return fallback
}

// warnIfChangelog prints the changelog warning when the statement is known to be
// non-append-only, and returns (isAppendOnly, appendOnlyKnown) for printQueryResult
// (which uses them to decide whether to show the Operation column).
func warnIfChangelog(result *query.Result) (bool, bool) {
	traits := result.Statement.Status.GetTraits()
	statementTraits := types.StatementTraits{FlinkGatewayV1StatementTraits: &traits}
	isAppendOnly, appendOnlyKnown := statementTraits.GetIsAppendOnly()
	if appendOnlyKnown && !isAppendOnly {
		output.ErrPrintln(false, "Warning: this statement emits updates and deletions. The rows below are the raw changelog, not a materialized table.")
	}
	return isAppendOnly, appendOnlyKnown
}

// phaseError reports a terminal phase that something other than this command
// caused (a server-side failure, or a stop/delete from elsewhere), or nil when
// the phase is not one of those.
func phaseError(name string, result *query.Result) error {
	switch result.Phase() {
	case types.FAILED:
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`statement "%s" failed: %s`, name, result.Statement.Status.GetDetail()),
			fmt.Sprintf("Inspect the failure with `confluent flink statement exception list %s`.", name),
		)
	case types.STOPPED:
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`statement "%s" was stopped before this command stopped it`, name),
			"The result set may be incomplete.",
		)
	case types.DELETING:
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`statement "%s" is being deleted`, name),
			"Its compute pool may have been deleted. The result set may be incomplete.",
		)
	}
	return nil
}

func (c *command) resolveDatabase(cmd *cobra.Command) (string, error) {
	database, err := cmd.Flags().GetString("database")
	if err != nil {
		return "", err
	}
	if database != "" {
		return database, nil
	}
	return c.Context.KafkaClusterContext.GetActiveKafkaClusterId(), nil
}

// resolveSQL reads the SQL from whichever of "sql"/"file" was given. Cobra's
// MarkFlagsOneRequired/MarkFlagsMutuallyExclusive only check whether a flag was
// set, not whether its value is non-empty, so `--sql ""` (e.g. from an unset
// shell variable) — or a --file pointing at an empty/whitespace-only file —
// passes flag-group validation; the emptiness check at the end covers both
// sources uniformly.
func resolveSQL(cmd *cobra.Command) (string, error) {
	sql, err := cmd.Flags().GetString("sql")
	if err != nil {
		return "", err
	}

	if sql == "" {
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return "", err
		}
		if file != "" {
			contents, err := os.ReadFile(file)
			if err != nil {
				return "", fmt.Errorf(`failed to read the SQL statement from "%s": %v`, file, err)
			}
			sql = string(contents)
		}
	}

	if strings.TrimSpace(sql) == "" {
		return "", errors.New("the SQL statement is required: pass it with `--sql` or `--file`")
	}
	return sql, nil
}

func (c *command) buildQueryProperties(cmd *cobra.Command, catalog, database string) (map[string]string, error) {
	statementProperties := map[string]string{
		config.KeyCatalog:    catalog,
		snapshotModeProperty: snapshotModeNow,
	}
	if database != "" {
		statementProperties[config.KeyDatabase] = database
	}

	configs, err := cmd.Flags().GetStringSlice("property")
	if err != nil {
		return nil, err
	}

	if len(configs) > 0 {
		configMap, err := properties.ConfigSliceToMap(configs)
		if err != nil {
			return nil, err
		}
		if mode, ok := configMap[snapshotModeProperty]; ok && mode != snapshotModeNow {
			return nil, errors.NewErrorWithSuggestions(
				fmt.Sprintf("the `--property` flag must not set `%s`", snapshotModeProperty),
				"This command only supports a snapshot (point-in-time, append-only) read, so this property is fixed and cannot be overridden.",
			)
		}
		for key, value := range configMap {
			statementProperties[key] = value
		}
	}

	return statementProperties, nil
}
