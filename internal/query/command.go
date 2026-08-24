package query

import (
	"context"
	goerrors "errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	cliconfig "github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	flinkerror "github.com/confluentinc/cli/v4/pkg/errors/flink"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/query"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/jwt"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/properties"
)

const (
	// snapshotModeProperty makes the statement a bounded, point-in-time read. It is
	// a statement property rather than a flag on the API, so the command sets it as
	// a default that `--property` can override.
	snapshotModeProperty = "sql.snapshot.mode"
	snapshotModeNow      = "now"

	// stopTimeout bounds how long we wait for a statement to be stopped after the
	// user interrupts the command.
	stopTimeout = 5 * time.Second

	// queryFeatureFlag gates the command's visibility.
	queryFeatureFlag = "cli.query"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

// New mounts `confluent query` at the top level rather than under `flink`: today it
// only ever talks to the Flink gateway, but the same one-shot-query ergonomics are
// meant to cover other backends (e.g. Lightning Tables) later without a rename.
func New(cfg *cliconfig.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query [name]",
		Short: "Run a bounded Flink SQL query and print its results.",
		Long: "Run a bounded (snapshot) Flink SQL query, block until it finishes, and print the complete result set.\n\n" +
			"Unlike statement creation, which submits a statement and returns immediately, this command waits for every " +
			"result page and exits with a non-zero status if the statement fails. It is intended for scripting and " +
			"one-shot queries against a bounded (point-in-time) result set.\n\n" +
			"With \"-o json\" or \"-o yaml\", output defaults to an envelope carrying the column schema alongside the rows, " +
			"since the rows on their own carry no type information. Pass \"--raw\" for a bare array of row objects instead.",
		Args: cobra.MaximumNArgs(1),
		// Hidden until the flag targets an org. cfg.IsTest keeps it visible to the
		// integration suite regardless of the (unreachable, in tests) LaunchDarkly
		// evaluation.
		Hidden: !(cfg.IsTest || featureflags.Manager.BoolVariation(queryFeatureFlag, cfg.Context(), cliconfig.CliLaunchDarklyClient, true, false)),
		Annotations: map[string]string{
			pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin,
		},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Run a bounded query in the current compute pool and print the rows as a table.",
				Code: `confluent query --sql "SELECT * FROM orders LIMIT 10;"`,
			},
			examples.Example{
				Text: "Run a bounded query against Kafka cluster \"my-cluster\" and emit JSON for a script to consume.",
				Code: `confluent query --sql "SELECT status, COUNT(*) FROM orders GROUP BY status;" --compute-pool lfcp-123456 --database my-cluster --output json`,
			},
			examples.Example{
				Text: "Emit a bare JSON array of rows, with no envelope, for a script that only wants the data.",
				Code: `confluent query --sql "SELECT * FROM orders LIMIT 10;" --output json --raw`,
			},
		),
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.runQuery

	cmd.Flags().String("sql", "", "The Flink SQL statement.")
	c.addComputePoolFlag(cmd)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	c.addDatabaseFlag(cmd)
	cmd.Flags().StringSlice("property", []string{}, "A mechanism to pass properties in the form key=value when creating a Flink statement.")
	cmd.Flags().Duration("timeout", config.DefaultTimeoutDuration, "Maximum time to wait for the query to finish before giving up.")
	cmd.Flags().Int("max-rows", 0, "Stop fetching and discard the rest after this many rows, or 0 to fetch every row. Client-side only: rows past the limit are still produced by the query.")
	cmd.Flags().Bool("raw", false, `Emit the rows as a bare array with no envelope. Requires "-o json" or "-o yaml".`)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("sql"))

	return cmd
}

// addComputePoolFlag and addDatabaseFlag mirror internal/flink's identical helpers.
// Duplicated rather than shared because this command intentionally lives outside the
// `flink` package boundary; hoisting them is worth doing only if a third caller shows up.
func (c *command) addComputePoolFlag(cmd *cobra.Command) {
	cmd.Flags().String("compute-pool", "", "Flink compute pool ID.")
	pcmd.RegisterFlagCompletionFunc(cmd, "compute-pool", c.autocompleteComputePools)
}

func (c *command) autocompleteComputePools(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	computePools, err := c.V2Client.ListFlinkComputePools("", environmentId, "")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(computePools))
	for i, computePool := range computePools {
		suggestions[i] = fmt.Sprintf("%s\t%s", computePool.GetId(), computePool.Spec.GetDisplayName())
	}
	return suggestions
}

func (c *command) addDatabaseFlag(cmd *cobra.Command) {
	cmd.Flags().String("database", "", "The database which will be used as the default database. When using Kafka, this is the cluster ID.")
	pcmd.RegisterFlagCompletionFunc(cmd, "database", c.autocompleteDatabases)
}

func (c *command) autocompleteDatabases(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	clusters, err := c.V2Client.ListKafkaClusters(environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.Spec.GetDisplayName())
	}
	return suggestions
}

type queryColumnOut struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
}

type queryOut struct {
	StatementName string           `json:"statement_name" yaml:"statement_name"`
	Phase         string           `json:"phase" yaml:"phase"`
	Columns       []queryColumnOut `json:"columns" yaml:"columns"`
	// Values carry their SQL type: a number serializes as a number and a NULL as null.
	// See types.StatementResultField.ToSerializedValue for what each type maps to.
	Rows      []map[string]any `json:"rows" yaml:"rows"`
	RowCount  int              `json:"row_count" yaml:"row_count"`
	Truncated bool             `json:"truncated" yaml:"truncated"`
	// Incomplete mirrors Result.Incomplete: the gateway stopped returning page tokens
	// while the statement was still running, so rows may be missing. The stderr
	// warning alone is invisible to a script that only reads stdout.
	Incomplete bool `json:"incomplete" yaml:"incomplete"`
}

func (c *command) runQuery(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	environment, _, err := c.V2Client.GetOrgEnvironment(environmentId)
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

	if computePool == "" && (cloud == "" || region == "") {
		return errors.New("the `--cloud` and `--region` flags are required when `--compute-pool` is not specified")
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

	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		return err
	}

	maxRows, err := cmd.Flags().GetInt("max-rows")
	if err != nil {
		return err
	}
	if maxRows < 0 {
		return errors.New("the `--max-rows` flag must not be negative")
	}

	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return err
	}
	if raw && !output.GetFormat(cmd).IsSerialized() {
		return errors.New("the `--raw` flag requires `-o json` or `-o yaml`")
	}

	statementProperties, err := c.buildQueryProperties(cmd, environment.GetDisplayName(), database)
	if err != nil {
		return err
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
	} else {
		client, err = c.GetFlinkGatewayClient(false)
	}
	if err != nil {
		return err
	}

	jwtValidator := jwt.NewValidator()

	serviceAccount, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}

	principal := serviceAccount
	if serviceAccount == "" {
		principal = c.Context.GetUser().GetResourceId()
	}

	if _, err := client.CreateStatement(statement, principal, environmentId, c.Context.LastOrgId); err != nil {
		return err
	}

	// From here on the statement exists server-side and is consuming pool capacity. A
	// job left running with nothing draining its collect-sink buffer stalls
	// indefinitely and keeps burning the compute it reserved — that was the truncation
	// bug: one of several exit paths simply forgot to stop it. Making that structurally
	// impossible, rather than remembering it at every exit, is the point of this defer:
	// it fires unless settled is true, and settled is only set once we know the
	// statement's fate — either it reached a terminal phase on its own, or something
	// already made one stop attempt on our behalf.
	settled := false
	defer func() {
		if !settled {
			c.stopStatement(client, environmentId, name)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	options := query.Options{
		Client:         client,
		EnvironmentId:  environmentId,
		OrganizationId: c.Context.LastOrgId,
		MaxRows:        maxRows,
		RequireBounded: true,
		RefreshToken:   c.refreshGatewayToken(client, jwtValidator),
	}

	result, err := query.Run(ctx, options, name)
	if err != nil {
		return c.handleQueryError(client, environmentId, name, err, &settled)
	}

	// The phase on result.Statement was read before the drain loop decided to
	// truncate or gave up on an exhausted page token — it says nothing about whether
	// the job kept running afterward. Treating it as proof of completion here would
	// be the same stale-read mistake the query package's drain() was fixed to avoid.
	// Truncated and Incomplete both mean "we chose to stop reading," which by the
	// same rule as the truncation bug always warrants a stop attempt.
	settled = !result.Truncated && !result.Incomplete && query.IsTerminal(result.Phase())

	if result.Phase() == types.FAILED {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`statement "%s" failed: %s`, name, result.Statement.Status.GetDetail()),
			fmt.Sprintf("Inspect the failure with `confluent flink statement exception list %s`.", name),
		)
	}

	if result.Incomplete {
		output.ErrPrintf(false, "Warning: the gateway stopped returning result pages while statement \"%s\" was still in phase %s. The result set below may be incomplete.\n", name, result.Phase())
	}

	if result.Truncated {
		output.ErrPrintf(false, "Warning: stopped after %d rows because of the \"--max-rows\" flag. The result set below is truncated.\n", maxRows)
	}

	traits := result.Statement.Status.GetTraits()
	statementTraits := types.StatementTraits{FlinkGatewayV1StatementTraits: &traits}
	isAppendOnly, appendOnlyKnown := statementTraits.GetIsAppendOnly()
	if appendOnlyKnown && !isAppendOnly {
		output.ErrPrintln(false, "Warning: this statement emits updates and deletions. The rows below are the raw changelog, not a materialized table.")
	}

	return c.printQueryResult(cmd, name, result, appendOnlyKnown && !isAppendOnly, raw)
}

// buildQueryProperties seeds the statement with the catalog and snapshot mode, then
// lets --property override anything, so the snapshot default never blocks a user who
// needs a different mode.
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
		for key, value := range configMap {
			statementProperties[key] = value
		}
	}

	return statementProperties, nil
}

// handleQueryError turns a failed or interrupted run into a message that always
// names the statement. settled is the same flag runQuery's deferred cleanup checks: a
// branch that stops the statement itself sets it to true so that deferred cleanup does
// not also attempt it, and a branch that has no better information about the
// statement's fate leaves it false so that cleanup does.
func (c *command) handleQueryError(client *ccloudv2.FlinkGatewayClient, environmentId, name string, err error, settled *bool) error {
	var unbounded *query.UnboundedError
	if goerrors.As(err, &unbounded) {
		// Only claim the statement was stopped when it actually was — stopStatement
		// already warns on its own when it could not.
		fate := fmt.Sprintf("Statement \"%s\" was stopped.", name)
		if !c.stopStatement(client, environmentId, name) {
			fate = fmt.Sprintf("Stop it with `confluent flink statement stop %s`.", name)
		}
		*settled = true
		return errors.NewErrorWithSuggestions(
			err.Error(),
			fmt.Sprintf("Bound the query with a LIMIT clause or a time predicate and run `confluent query` again. %s", fate),
		)
	}

	if goerrors.Is(err, context.Canceled) || goerrors.Is(err, context.DeadlineExceeded) {
		c.stopStatement(client, environmentId, name)
		*settled = true
		reason := "interrupted"
		if goerrors.Is(err, context.DeadlineExceeded) {
			reason = "timed out"
		}
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`query %s before statement "%s" finished`, reason, name),
			fmt.Sprintf("Check the statement with `confluent flink statement describe %s`, or raise the limit with the `--timeout` flag.", name),
		)
	}

	// Every other error, including ResultsFetchError below, leaves settled false: we
	// have no positive signal the statement is done, so the caller's deferred cleanup
	// makes the one stop attempt these branches used to skip.
	var resultsFetchErr *query.ResultsFetchError
	if goerrors.As(err, &resultsFetchErr) {
		// A 404 here is ambiguous — it is also what a deleted or mistyped statement
		// would return — so the suggestion names both possibilities rather than
		// asserting the page expired. The gateway's actual retention window is not
		// yet confirmed (see the query package README); this only makes the raw
		// error actionable, it does not resolve that open question.
		var coder flinkerror.Coder
		if goerrors.As(err, &coder) && coder.StatusCode() == http.StatusNotFound {
			return errors.NewErrorWithSuggestions(
				resultsFetchErr.Error(),
				fmt.Sprintf("The gateway only retains result pages for a limited window; this can happen if the query ran long enough for an earlier page to expire, or if the statement no longer exists. Re-run the query, or check `confluent flink statement describe %s`.", name),
			)
		}
		return errors.NewErrorWithSuggestions(
			resultsFetchErr.Error(),
			fmt.Sprintf("Check the statement with `confluent flink statement describe %s`.", name),
		)
	}

	return errors.NewErrorWithSuggestions(
		err.Error(),
		fmt.Sprintf("Check the statement with `confluent flink statement describe %s`.", name),
	)
}

// refreshGatewayToken mirrors the check the interactive shell makes before every
// gateway call: the dataplane token behind client is short-lived, and this command has
// no equivalent of the shell's synchronizedTokenRefresh wrapping every call, so without
// this a query that outlives that token dies on a 401 well before the command's own
// --timeout is reached.
func (c *command) refreshGatewayToken(client *ccloudv2.FlinkGatewayClient, jwtValidator jwt.Validator) func() error {
	return func() error {
		jwtCtx := &cliconfig.Context{State: &cliconfig.ContextState{AuthToken: client.AuthToken}}
		if jwtValidator.Validate(jwtCtx) == nil {
			return nil
		}

		dataplaneToken, err := auth.GetDataplaneToken(c.Context)
		if err != nil {
			return err
		}
		client.AuthToken = dataplaneToken
		return nil
	}
}

// stopStatement makes a best-effort, bounded attempt to stop a statement we are
// abandoning, and reports the outcome either way. Both outcomes matter to the user:
// a successful stop frees pool capacity, a failed one means there is a statement
// still running under a name they need.
func (c *command) stopStatement(client *ccloudv2.FlinkGatewayClient, environmentId, name string) bool {
	done := make(chan error, 1)
	go func() {
		// The gateway replaces the statement wholesale rather than patching it, so a
		// body carrying only spec.stopped is rejected as malformed and the statement
		// keeps running. Read the statement back and flip the flag on what it returns.
		statement, err := client.GetStatement(environmentId, name, c.Context.LastOrgId)
		if err != nil {
			done <- err
			return
		}
		statement.Spec.Stopped = flinkgatewayv1.PtrBool(true)
		done <- client.UpdateStatement(environmentId, name, c.Context.LastOrgId, statement)
	}()

	select {
	case err := <-done:
		if err != nil {
			output.ErrPrintf(false, "Warning: could not stop statement \"%s\": %v. It may still be running.\n", name, err)
			return false
		}
		output.ErrPrintf(false, "Stopped statement \"%s\".\n", name)
		return true
	case <-time.After(stopTimeout):
		output.ErrPrintf(false, "Warning: timed out trying to stop statement \"%s\". It may still be running.\n", name)
		return false
	}
}

func (c *command) printQueryResult(cmd *cobra.Command, name string, result *query.Result, showOperation, raw bool) error {
	columns := make([]queryColumnOut, len(result.Columns))
	headers := make([]string, len(result.Columns))
	for i, column := range result.Columns {
		columnType := column.GetType()
		columns[i] = queryColumnOut{Name: column.GetName(), Type: columnType.GetType()}
		headers[i] = column.GetName()
	}

	if output.GetFormat(cmd).IsSerialized() {
		rows := make([]map[string]any, len(result.Rows))
		for i, row := range result.Rows {
			fields := make(map[string]any, len(headers))
			for j, field := range row.GetFields() {
				if j < len(headers) {
					fields[headers[j]] = field.ToSerializedValue()
				}
			}
			rows[i] = fields
		}

		// R3 asks for a bare array of row objects; R10 asks for an ordered, typed
		// schema. A bare array has nowhere to put the schema, so the envelope is the
		// default and --raw opts into the bare array R3 describes.
		if raw {
			return output.SerializedOutput(cmd, rows)
		}

		return output.SerializedOutput(cmd, &queryOut{
			StatementName: name,
			Phase:         string(result.Phase()),
			Columns:       columns,
			Rows:          rows,
			RowCount:      len(rows),
			Truncated:     result.Truncated,
			Incomplete:    result.Incomplete,
		})
	}

	if len(headers) == 0 || len(result.Rows) == 0 {
		output.ErrPrintf(false, "The query returned no rows. Statement \"%s\" is in phase %s.\n", name, result.Phase())
		return nil
	}

	if showOperation {
		headers = append([]string{"Operation"}, headers...)
	}

	rows := make([][]string, len(result.Rows))
	for i, row := range result.Rows {
		fields := make([]string, 0, len(headers))
		if showOperation {
			fields = append(fields, row.Operation.String())
		}
		for _, field := range row.GetFields() {
			fields = append(fields, field.ToString())
		}
		rows[i] = fields
	}

	// Deliberately no column truncation: the shell truncates to fit a terminal, but
	// this command is expected to be piped, and silently shortening a value would
	// corrupt whatever reads it.
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(headers)
	table.AppendBulk(rows)
	table.Render()

	return nil
}
