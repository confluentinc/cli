package flink

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

	// flinkQueryFeatureFlag gates the command's visibility. It lives in the Confluent
	// Cloud LaunchDarkly project (like the other flink.* flags this CLI reads), so it
	// can be targeted by org.
	flinkQueryFeatureFlag = "flink.query.cli.enable"
)

type queryCommand struct {
	*pcmd.AuthenticatedCLICommand

	// authTokenMu guards client.AuthToken against a leaked refresh racing a stop attempt.
	authTokenMu sync.Mutex
}

func (*command) newQueryCommand(cfg *cliconfig.Config, prerunner pcmd.PreRunner) *cobra.Command {
	c := &queryCommand{}
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Run a bounded Flink SQL query and print its results.",
		Long: "Run a bounded Flink SQL query, wait for it to finish, and print the results.\n\n" +
			"Provide the SQL statement with `--sql`, or use `--file` to read it from a file.\n\n" +
			"Unlike creating a statement, which returns a handle as soon as it's submitted (optionally waiting only " +
			"until it starts or fails, with `--wait`) and never the rows, this command always waits for the statement " +
			"to finish, exiting non-zero if it fails. Use it for scripts and one-time queries against a bounded, " +
			"point-in-time result set.\n\n" +
			"With `-o human` (the default), only the first 100 rows are printed as a preview. Raise `--max-rows` to fetch " +
			"more, or use `-o json` / `-o yaml` for the complete result set.\n\n" +
			"`-o json` and `-o yaml` return a bare array of row objects.\n\n" +
			"When `--max-rows` cuts a result short, a warning is printed to standard error and the command still exits 0, so a script reading the rows on standard output is unaffected.",
		Args: cobra.NoArgs,
		// Hidden until the flag targets an org; cfg.IsTest keeps it visible to the
		// integration suite regardless of the (unreachable in tests) LD evaluation.
		Hidden: !(cfg.IsTest || featureflags.Manager.BoolVariation(flinkQueryFeatureFlag, cfg.Context(), cliconfig.CcloudProdLaunchDarklyClient, true, false)),
		Annotations: map[string]string{
			pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin,
		},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Run a bounded query in the current compute pool and print the rows as a table.",
				Code: `confluent flink query --sql "SELECT * FROM orders LIMIT 10;"`,
			},
			examples.Example{
				Text: "Run a bounded query against Kafka cluster \"my-cluster\" and return JSON for a script.",
				Code: `confluent flink query --sql "SELECT status, COUNT(*) FROM orders GROUP BY status;" --compute-pool lfcp-123456 --database my-cluster --output json`,
			},
		),
		RunE: c.runQuery,
	}
	c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)

	cmd.Flags().String("sql", "", `Flink SQL statement. Alternatively, use "--file" or "-f".`)
	cmd.Flags().StringP("file", "f", "", "Path to a file that contains the Flink SQL statement. Alternatively, use --sql.")
	pcmd.AddComputePoolFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddDatabaseFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().StringSlice("property", []string{}, "Properties for the Flink statement in key=value format.")
	cmd.Flags().Duration("timeout", config.DefaultTimeoutDuration, "Maximum time to wait for the query to finish.")
	cmd.Flags().Int("max-rows", 0, `Maximum number of rows to fetch. Defaults to 100 for "-o human"; raise it to fetch more. "-o json"/"-o yaml" fetch every row by default. This limit is client-side only; the query still produces rows after the limit is reached.`)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)

	cmd.MarkFlagsOneRequired("sql", "file")
	cmd.MarkFlagsMutuallyExclusive("sql", "file")

	return cmd
}

// defaultHumanRows caps the -o human preview when --max-rows isn't set; json/yaml
// stay uncapped. Like Spark's df.show() default.
const defaultHumanRows = 100

func (c *queryCommand) runQuery(cmd *cobra.Command, _ []string) error {
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

	principal, err := c.resolvePrincipal(cmd)
	if err != nil {
		return err
	}

	timeout, maxRows, err := resolveQueryFlags(cmd)
	if err != nil {
		return err
	}

	maxRows, humanDefaultCap := resolveDisplayCap(output.GetFormat(cmd), cmd.Flags().Changed("max-rows"), maxRows)

	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	client, name, err := c.createQueryStatement(ctx, cmd, environmentId, database, sql, principal)
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
	defer func() { c.releaseStatement(client, environmentId, name, settled, announceStop) }()

	options := query.Options{
		Client:         client,
		EnvironmentId:  environmentId,
		OrganizationId: c.Context.LastOrgId,
		MaxRows:        maxRows,
		RequireBounded: true,
		RefreshToken:   c.refreshGatewayToken(client, jwt.NewValidator()),
	}

	// json/yaml stream page by page so a large result never buffers; -o human is
	// buffered instead (it needs the full set to align columns). On a mid-stream
	// failure the emitted json/yaml is left unterminated.
	var streamer resultStreamer
	if output.GetFormat(cmd).IsSerialized() {
		streamer = newResultStreamer(output.GetFormat(cmd))
		options.OnSchema = streamer.setColumns
		options.OnRows = streamer.writeRows
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
	// The implicit human cap isn't user-requested, so its cleanup stays quiet.
	announceStop = result.Truncated && !humanDefaultCap
	return c.emitResult(name, result, streamer, maxRows, humanDefaultCap, &announceStop)
}

// emitResult finishes output after a successful drain. Streamed formats already
// wrote rows during Run, so this closes the document and surfaces any terminal-
// phase error afterward; -o human renders the buffered table. It sets
// *announceStop on a write failure so the deferred cleanup speaks up.
func (c *queryCommand) emitResult(name string, result *query.Result, streamer resultStreamer, maxRows int, humanDefaultCap bool, announceStop *bool) error {
	// -o human hasn't printed anything yet, so a terminal-phase error (STOPPED/
	// DELETING) returns before the table. Streamed rows already reached stdout, so
	// fall through and let close() terminate the document first.
	phaseErr := phaseError(name, result)
	if phaseErr != nil && streamer == nil {
		return phaseErr
	}

	if result.Truncated && !humanDefaultCap {
		output.ErrPrintf(false, "Warning: stopped after %d rows because of the `--max-rows` flag. The result set below is truncated.\n", maxRows)
	}
	isAppendOnly, appendOnlyKnown := warnIfChangelog(result)

	if streamer != nil {
		if err := streamer.close(); err != nil {
			*announceStop = true
			return err
		}
		return phaseErr
	}

	if err := printHumanResult(os.Stdout, name, result, isAppendOnly, appendOnlyKnown); err != nil {
		*announceStop = true
		return err
	}

	// After the table, on stderr (so a piped table stays clean), only when the
	// preview cap actually cut rows off.
	if result.Truncated && humanDefaultCap {
		output.ErrPrintf(false, "\nOnly showing the first %d rows. Use `--max-rows` to show more, or `-o json`/`-o yaml` for the full result.\n", maxRows)
	}
	return nil
}

// releaseStatement stops the leftover statement on exit unless the run already
// settled it. announceStop picks a reported vs quiet stop; a quiet stop still
// surfaces a failure, since a running statement keeps burning compute.
func (c *queryCommand) releaseStatement(client *ccloudv2.FlinkGatewayClient, environmentId, name string, settled, announceStop bool) {
	if settled {
		return
	}
	if announceStop {
		c.stopStatementAndReport(client, environmentId, name)
		return
	}
	if ok, err := c.stopStatement(client, environmentId, name); !ok {
		reportStopFailure(name, err)
	}
}

// resolveDisplayCap defaults -o human to defaultHumanRows when --max-rows isn't
// set (json/yaml stay uncapped), and reports whether that implicit cap applied so
// the caller shows the preview notice rather than the --max-rows warning.
func resolveDisplayCap(format output.Format, maxRowsChanged bool, maxRows int) (int, bool) {
	if format == output.Human && !maxRowsChanged {
		return defaultHumanRows, true
	}
	return maxRows, false
}

// resolveQueryFlags reads and validates the numeric/output flags that gate the run
// before any network call: --timeout must be positive, and --max-rows must not be
// negative (0 means no limit).
func resolveQueryFlags(cmd *cobra.Command) (time.Duration, int, error) {
	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		return 0, 0, err
	}
	if timeout <= 0 {
		return 0, 0, errors.NewErrorWithSuggestions(
			"the `--timeout` flag must be positive",
			"Set `--timeout` to a positive duration, such as `30s` or `10m`.",
		)
	}

	maxRows, err := cmd.Flags().GetInt("max-rows")
	if err != nil {
		return 0, 0, err
	}
	if maxRows < 0 {
		return 0, 0, errors.NewErrorWithSuggestions(
			"the `--max-rows` flag must not be negative",
			"Set `--max-rows` to 0 for no limit, or a positive integer.",
		)
	}

	return timeout, maxRows, nil
}

// createQueryStatement resolves the environment and gateway client, builds the
// bounded snapshot statement, and submits it, returning the client and the
// generated statement name. All command flags are resolved by the caller and
// passed in, so this makes only API calls. A cancellation during any of these
// pre-result steps is mapped to interruptedError.
func (c *queryCommand) createQueryStatement(ctx context.Context, cmd *cobra.Command, environmentId, database, sql, principal string) (*ccloudv2.FlinkGatewayClient, string, error) {
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
func (c *queryCommand) resolvePrincipal(cmd *cobra.Command) (string, error) {
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
func (c *queryCommand) interruptOr(cmd *cobra.Command, err error, name string, stopped bool, fallback error) error {
	if isInterrupted(err) {
		return interruptedError(cmd, err, name, stopped)
	}
	return fallback
}

// warnIfChangelog prints the changelog warning when the statement is known to be
// non-append-only, and returns (isAppendOnly, appendOnlyKnown) for printHumanResult
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

func (c *queryCommand) resolveDatabase(cmd *cobra.Command) (string, error) {
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

func (c *queryCommand) buildQueryProperties(cmd *cobra.Command, catalog, database string) (map[string]string, error) {
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
