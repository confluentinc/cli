package lightning

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
)

// defaultHumanRows is the default -o human row cap, as in `confluent flink query`.
const defaultHumanRows = 100

type queryCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newQueryCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := &queryCommand{}
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Run a SQL query on Lightning Tables and print its results.",
		Long: "Run a SQL query on Lightning Tables and print the results.\n\n" +
			"Provide the SQL statement with `--sql`, or use `--file` to read the statement from a file. The query runs against the Kafka cluster set with `--cluster`, or the active Kafka cluster if the flag is not set.\n\n" +
			"The query runs synchronously. If the result is too large or takes too long to return, the command fails instead of starting an asynchronous query; narrow the query, for example with a \"LIMIT\" clause or a more selective \"WHERE\" clause.\n\n" +
			"With `-o human` (the default), only the first 100 rows are printed as a preview. Raise `--max-rows` to fetch more, or use `-o json` or `-o yaml` for the complete result set.\n\n" +
			"The `-o json` and `-o yaml` flags return a bare array of row objects.\n\n" +
			"When `--max-rows` cuts a result short, a warning is printed to standard error and the command still exits 0, so a script reading the rows on standard output is unaffected.",
		Args: cobra.NoArgs,
		Annotations: map[string]string{
			pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin,
		},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Run a query against the active Kafka cluster and print the rows as a table.",
				Code: `confluent lightning query --sql "SELECT * FROM orders WHERE status = 'completed'"`,
			},
			examples.Example{
				Text: `Run a query against Kafka cluster "lkc-123456" and return JSON for a script.`,
				Code: `confluent lightning query --sql "SELECT order_id, status FROM orders LIMIT 10" --cluster lkc-123456 --output json`,
			},
		),
		RunE: c.query,
	}
	c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)

	cmd.Flags().String("sql", "", `SQL statement. Alternatively, use "--file" or "-f".`)
	cmd.Flags().StringP("file", "f", "", `Path to a file that contains the SQL statement. Alternatively, use "--sql".`)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().Int("max-rows", 0, `Maximum number of rows to fetch, or 0 for no limit. Defaults to 100 for "-o human"; raise it to fetch more. "-o json"/"-o yaml" fetch every row by default.`)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("sql", "file")
	cmd.MarkFlagsMutuallyExclusive("sql", "file")

	return cmd
}

func (c *queryCommand) query(cmd *cobra.Command, _ []string) error {
	sql, err := resolveSQL(cmd)
	if err != nil {
		return err
	}

	maxRows, err := resolveMaxRows(cmd)
	if err != nil {
		return err
	}
	maxRows, humanDefaultCap := resolveDisplayCap(output.GetFormat(cmd), cmd.Flags().Changed("max-rows"), maxRows)

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	clusterId := c.Context.KafkaClusterContext.GetActiveKafkaClusterId()
	if clusterId == "" {
		return errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaSelectedSuggestions)
	}

	client, err := c.getLightningClient(cmd, environmentId, clusterId)
	if err != nil {
		return err
	}

	result, err := client.ExecuteQuery(environmentId, clusterId, sql, maxResultRows(maxRows))
	if err != nil {
		return err
	}

	headers, rows, err := convertResult(result)
	if err != nil {
		return err
	}

	return printResult(os.Stdout, os.Stderr, output.GetFormat(cmd), applyRowLimits(headers, rows, maxRows, humanDefaultCap))
}

// getLightningClient returns a client for the Lightning endpoint in the Kafka cluster's cloud and region.
func (c *queryCommand) getLightningClient(cmd *cobra.Command, environmentId, clusterId string) (*ccloudv2.LightningClient, error) {
	cluster, httpResp, err := c.V2Client.DescribeKafkaCluster(clusterId, environmentId)
	if err != nil {
		return nil, errors.CatchKafkaNotFoundError(err, clusterId, httpResp)
	}

	url, err := ccloudv2.GetLightningUrl(c.Config, cluster.Spec.GetCloud(), cluster.Spec.GetRegion())
	if err != nil {
		return nil, err
	}

	unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
	if err != nil {
		return nil, err
	}

	dataplaneToken, err := auth.GetDataplaneToken(c.Context)
	if err != nil {
		return nil, err
	}

	log.CliLogger.Debugf("The final url used for setting up Lightning dataplane client is: %s\n", url)
	return ccloudv2.NewLightningClient(url, c.Version.UserAgent, unsafeTrace, dataplaneToken), nil
}

func resolveMaxRows(cmd *cobra.Command) (int, error) {
	maxRows, err := cmd.Flags().GetInt("max-rows")
	if err != nil {
		return 0, err
	}
	if maxRows < 0 {
		return 0, errors.NewErrorWithSuggestions(
			"the `--max-rows` flag must not be negative",
			"Set `--max-rows` to 0 for no limit, or a positive integer.",
		)
	}
	return maxRows, nil
}

// resolveDisplayCap caps -o human at defaultHumanRows unless --max-rows is set, and reports whether it did.
func resolveDisplayCap(format output.Format, maxRowsChanged bool, maxRows int) (int, bool) {
	if format == output.Human && !maxRowsChanged {
		return defaultHumanRows, true
	}
	return maxRows, false
}

// maxResultRows asks for one row more than the cap to detect truncation, or 0 for no limit.
func maxResultRows(maxRows int) int32 {
	if maxRows <= 0 {
		return 0
	}
	return int32(min(maxRows, math.MaxInt32-1) + 1)
}

type queryResult struct {
	headers         []string
	rows            [][]types.StatementResultField
	maxRows         int  // 0 means no cap
	truncated       bool // maxRows cut rows off
	humanDefaultCap bool // maxRows is the -o human default, not --max-rows
}

func applyRowLimits(headers []string, rows [][]types.StatementResultField, maxRows int, humanDefaultCap bool) queryResult {
	result := queryResult{headers: headers, rows: rows, maxRows: maxRows, humanDefaultCap: humanDefaultCap}
	if maxRows > 0 && len(rows) > maxRows {
		result.rows = rows[:maxRows]
		result.truncated = true
	}
	return result
}

// resolveSQL reads --sql or --file and rejects an empty statement, which Cobra's flag checks allow.
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
				return "", errors.NewErrorWithSuggestions(
					fmt.Sprintf(`failed to read the SQL statement from "%s": %v`, file, err),
					"Check that the `--file` path is correct and the file is readable.",
				)
			}
			sql = string(contents)
		}
	}

	// Some editors start a file with a byte order mark, which isn't part of the SQL.
	sql = strings.TrimPrefix(sql, "\uFEFF")

	if strings.TrimSpace(sql) == "" {
		return "", errors.NewErrorWithSuggestions(
			"the SQL statement is empty",
			"Pass the statement with `--sql`, or point `--file` at a file that contains one.",
		)
	}
	return sql, nil
}
