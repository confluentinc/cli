package query

import (
	"context"
	goerrors "errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

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
	"github.com/confluentinc/cli/v4/pkg/wait"
)

const (
	// snapshotModeProperty makes this a bounded, point-in-time read; not overridable via --property.
	snapshotModeProperty = "sql.snapshot.mode"
	snapshotModeNow      = "now"

	// engineSnapshot is the only value Engine ever takes today; see queryOut.Engine.
	engineSnapshot = "snapshot"

	// stopTimeout bounds how long we wait for a statement to stop after interrupt.
	stopTimeout = 5 * time.Second

	// createStatementGracePeriod bounds how long we wait for a cancelled CreateStatement to land, so it can still be stopped.
	createStatementGracePeriod = 10 * time.Second

	// queryFeatureFlag gates the command's visibility.
	queryFeatureFlag = "cli.query"
)

type command struct {
	*pcmd.AuthenticatedCLICommand

	// authTokenMu guards client.AuthToken against a leaked refresh racing a stop attempt.
	authTokenMu sync.Mutex
}

// New mounts `confluent query` at the top level, not under `flink`, to also cover
// future backends like Lightning Tables without a rename.
func New(cfg *cliconfig.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query [sql]",
		Short: "Run a bounded Flink SQL query and print its results.",
		Long: "Run a bounded (snapshot) Flink SQL query, block until it finishes, and print the complete result set.\n\n" +
			"The SQL can be given as `--sql` or as a positional argument, but not both.\n\n" +
			"Unlike statement creation, which submits a statement and returns immediately, this command waits for every " +
			"result page and exits with a non-zero status if the statement fails. It is intended for scripting and " +
			"one-shot queries against a bounded (point-in-time) result set.\n\n" +
			"With `-o json` or `-o yaml`, output defaults to an envelope carrying the column schema alongside the rows, " +
			"since the rows on their own carry no type information. Pass `--raw` for a bare array of row objects instead.",
		Args: cobra.MaximumNArgs(1),
		// Hidden until the flag targets an org; cfg.IsTest keeps it visible to the
		// integration suite regardless of the (unreachable in tests) LD evaluation.
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
			examples.Example{
				Text: "Pass the SQL as a positional argument instead of `--sql`.",
				Code: `confluent query "SELECT * FROM orders LIMIT 10;"`,
			},
		),
	}

	c := &command{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.runQuery

	cmd.Flags().String("sql", "", `The Flink SQL statement. Alternatively, pass it as a positional argument or with "-f".`)
	cmd.Flags().StringP("file", "f", "", `Path to a file containing the Flink SQL statement. Alternatively, pass the SQL with "--sql" or as a positional argument.`)
	c.addComputePoolFlag(cmd)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	c.addDatabaseFlag(cmd)
	c.addClusterAlias(cmd)
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

	return cmd
}

// addComputePoolFlag and addDatabaseFlag mirror internal/flink's helpers, duplicated
// since this command lives outside the `flink` package boundary.
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

// addClusterAlias is a separate flag, not shared storage: ParseFlagsIntoContext
// persists "cluster" to the active Kafka context but never "database".
func (c *command) addClusterAlias(cmd *cobra.Command) {
	cmd.Flags().String("cluster", "", `Alias for "--database". Unlike "--database", this also sets the CLI's active Kafka cluster context, the same as it does on every other command.`)
	pcmd.RegisterFlagCompletionFunc(cmd, "cluster", c.autocompleteDatabases)
}

// addCatalogAlias shares --environment's pflag.Value directly, since --environment
// already persists to context and ParseFlagsIntoContext reads it before RunE runs.
func (c *command) addCatalogAlias(cmd *cobra.Command) {
	environmentFlag := cmd.Flags().Lookup("environment")
	cmd.Flags().Var(environmentFlag.Value, "catalog", `Alias for "--environment".`)
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
	StatementName string `json:"statement_name" yaml:"statement_name"`
	// Engine is always "snapshot" until M2 (Lightning routing) ships; emitting it
	// now means a script switching on this field today won't need to change later.
	Engine  string           `json:"engine" yaml:"engine"`
	Phase   string           `json:"phase" yaml:"phase"`
	Columns []queryColumnOut `json:"columns" yaml:"columns"`
	// Values carry their SQL type (number, null, etc); see
	// types.StatementResultField.ToSerializedValue.
	Rows      []map[string]any `json:"rows" yaml:"rows"`
	RowCount  int              `json:"row_count" yaml:"row_count"`
	Truncated bool             `json:"truncated" yaml:"truncated"`
	// AppendOnly is nil until traits are known; false means Rows is a changelog, not a materialized table.
	AppendOnly *bool `json:"append_only,omitempty" yaml:"append_only,omitempty"`
}

func (c *command) runQuery(cmd *cobra.Command, args []string) error {
	if err := resolveEnvironmentAlias(cmd); err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	sql, err := resolveSQL(cmd, args)
	if err != nil {
		return err
	}

	database, err := c.resolveDatabase(cmd)
	if err != nil {
		return err
	}

	timeout, err := cmd.Flags().GetDuration("wait-timeout")
	if err != nil {
		return err
	}
	if timeout <= 0 {
		return errors.New("the `--wait-timeout` flag must be positive")
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

	// Built before any network call so --wait-timeout bounds setup too, not just
	// draining; every call below is wrapped in wait.Call since the SDK clients ignore context.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	environment, err := wait.Call(ctx, func() (orgv2.OrgV2Environment, error) {
		env, _, err := c.V2Client.GetOrgEnvironment(environmentId)
		return env, err
	})
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), "List available environments with `confluent environment list`.")
	}

	// computePool falls back to context (`flink compute-pool use`);
	// GetFlinkGatewayClient below decides whether pool or cloud/region is enough.
	computePool := c.Context.GetCurrentFlinkComputePool()

	name := types.GenerateStatementName()

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
		client, err = wait.Call(ctx, func() (*ccloudv2.FlinkGatewayClient, error) { return c.GetFlinkGatewayClient(true) })
	} else {
		client, err = wait.Call(ctx, func() (*ccloudv2.FlinkGatewayClient, error) { return c.GetFlinkGatewayClient(false) })
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

	if err := createStatement(ctx, createStatementGracePeriod, func() (flinkgatewayv1.SqlV1Statement, error) {
		return client.CreateStatement(statement, principal, environmentId, c.Context.LastOrgId)
	}, func() { c.stopStatement(client, environmentId, name) }); err != nil {
		return err
	}

	// The statement now exists server-side; this defer stops it unless settled is
	// set, so no exit path can forget to release the compute it's holding.
	settled := false
	defer func() {
		if !settled {
			c.stopStatement(client, environmentId, name)
		}
	}()

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

	// drain() refreshes result.Statement, so this reflects reality even after
	// Truncated — a job can stay RUNNING after its last row ships either way.
	settled = query.IsTerminal(result.Phase())

	// STOPPED/DELETING here means something other than us ended the statement.
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

	if result.Truncated {
		output.ErrPrintf(false, "Warning: stopped after %d rows because of the `--max-rows` flag. The result set below is truncated.\n", maxRows)
	}

	traits := result.Statement.Status.GetTraits()
	statementTraits := types.StatementTraits{FlinkGatewayV1StatementTraits: &traits}
	isAppendOnly, appendOnlyKnown := statementTraits.GetIsAppendOnly()
	if appendOnlyKnown && !isAppendOnly {
		output.ErrPrintln(false, "Warning: this statement emits updates and deletions. The rows below are the raw changelog, not a materialized table.")
	}

	return c.printQueryResult(cmd, name, result, isAppendOnly, appendOnlyKnown, raw)
}

func resolveEnvironmentAlias(cmd *cobra.Command) error {
	if cmd.Flags().Changed("environment") && cmd.Flags().Changed("catalog") {
		return errors.New("the environment must not be given both with the `--environment` flag and with the `--catalog` flag")
	}
	return nil
}

func (c *command) resolveDatabase(cmd *cobra.Command) (string, error) {
	database, err := cmd.Flags().GetString("database")
	if err != nil {
		return "", err
	}
	cluster, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return "", err
	}
	if database != "" && cluster != "" {
		return "", errors.New("the database must not be given both with the `--database` flag and with the `--cluster` flag")
	}
	if cluster != "" {
		return cluster, nil
	}
	if database != "" {
		return database, nil
	}
	return c.Context.KafkaClusterContext.GetActiveKafkaClusterId(), nil
}

func resolveSQL(cmd *cobra.Command, args []string) (string, error) {
	sql, err := cmd.Flags().GetString("sql")
	if err != nil {
		return "", err
	}

	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return "", err
	}

	positional := len(args) == 1

	switch {
	case positional && sql != "", positional && file != "", sql != "" && file != "":
		return "", errors.New("the SQL statement must be given exactly one way: as a positional argument, with the `--sql` flag, or with the `--file` flag")
	case positional:
		return args[0], nil
	case sql != "":
		return sql, nil
	case file != "":
		contents, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf(`failed to read the SQL statement from "%s": %v`, file, err)
		}
		return string(contents), nil
	default:
		return "", errors.New("the SQL statement is required: pass it as a positional argument, with the `--sql` flag, or with the `--file` flag")
	}
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

// handleQueryError turns a failed or interrupted run into a message naming the
// statement; settled marks whether this call already stopped it.
func (c *command) handleQueryError(client *ccloudv2.FlinkGatewayClient, environmentId, name string, err error, settled *bool) error {
	var unbounded *query.UnboundedError
	if goerrors.As(err, &unbounded) {
		// Only claim the statement was stopped when it actually was.
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
			fmt.Sprintf("Check the statement with `confluent flink statement describe %s`, or raise the limit with the `--wait-timeout` flag.", name),
		)
	}

	// Every other error, including ResultsFetchError below, leaves settled false:
	// the caller's deferred cleanup makes the stop attempt these branches skip.
	var resultsFetchErr *query.ResultsFetchError
	if goerrors.As(err, &resultsFetchErr) {
		var coder flinkerror.Coder
		if goerrors.As(err, &coder) {
			switch coder.StatusCode() {
			case http.StatusNotFound:
				// A 404 means the statement is gone or mistyped; an expired result window is a separate 408 (below).
				return errors.NewErrorWithSuggestions(
					resultsFetchErr.Error(),
					fmt.Sprintf("Statement \"%s\" no longer exists — it may have been deleted, or the name is mistyped. Check `confluent flink statement describe %s`.", name, name),
				)
			case http.StatusRequestTimeout:
				// Results are retained for one hour; past that the gateway returns 408 with its own message.
				return errors.NewErrorWithSuggestions(
					resultsFetchErr.Error(),
					"Re-run the query — the result window for this statement has closed.",
				)
			}
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

// refreshGatewayToken mirrors the shell's pre-call check: without it, a query
// outliving the short-lived dataplane token dies on a 401 before --wait-timeout.
func (c *command) refreshGatewayToken(client *ccloudv2.FlinkGatewayClient, jwtValidator jwt.Validator) func() error {
	return func() error {
		c.authTokenMu.Lock()
		defer c.authTokenMu.Unlock()

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

// createStatement runs create; on cancellation it waits up to gracePeriod for it to
// land and calls cleanup instead of leaking an unstoppable statement.
func createStatement(ctx context.Context, gracePeriod time.Duration, create func() (flinkgatewayv1.SqlV1Statement, error), cleanup func()) error {
	done := make(chan error, 1)
	go func() {
		_, err := create()
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		select {
		case err := <-done:
			if err == nil {
				cleanup()
			}
		case <-time.After(gracePeriod):
		}
		return ctx.Err()
	}
}

// stopStatement makes a best-effort, bounded attempt to stop an abandoned
// statement and reports the outcome either way.
func (c *command) stopStatement(client *ccloudv2.FlinkGatewayClient, environmentId, name string) bool {
	done := make(chan error, 1)
	go func() {
		c.authTokenMu.Lock()
		defer c.authTokenMu.Unlock()
		done <- client.StopStatement(environmentId, name, c.Context.LastOrgId)
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

func (c *command) printQueryResult(cmd *cobra.Command, name string, result *query.Result, isAppendOnly, appendOnlyKnown, raw bool) error {
	columns := make([]queryColumnOut, len(result.Columns))
	headers := make([]string, len(result.Columns))
	for i, column := range result.Columns {
		columnType := column.GetType()
		columns[i] = queryColumnOut{Name: column.GetName(), Type: columnType.GetType()}
		headers[i] = column.GetName()
	}

	showOperation := appendOnlyKnown && !isAppendOnly

	if output.GetFormat(cmd).IsSerialized() {
		rows := make([]map[string]any, len(result.Rows))
		for i, row := range result.Rows {
			// Every row is guaranteed len(headers) fields: Run() hard-errors on a
			// row/schema mismatch before this function ever sees a result.
			fields := make(map[string]any, len(headers))
			for j, field := range row.GetFields() {
				fields[headers[j]] = field.ToSerializedValue()
			}
			rows[i] = fields
		}

		// A bare array has nowhere to put the schema, so the envelope is the
		// default and --raw opts into the bare array.
		if raw {
			return output.SerializedOutput(cmd, rows)
		}

		var appendOnly *bool
		if appendOnlyKnown {
			appendOnly = &isAppendOnly
		}

		return output.SerializedOutput(cmd, &queryOut{
			StatementName: name,
			Engine:        engineSnapshot,
			Phase:         string(result.Phase()),
			Columns:       columns,
			Rows:          rows,
			RowCount:      len(rows),
			Truncated:     result.Truncated,
			AppendOnly:    appendOnly,
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

	// No column truncation: this command is expected to be piped, and shortening
	// a value would corrupt whatever reads it.
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(headers)
	table.AppendBulk(rows)
	table.Render()

	return nil
}
