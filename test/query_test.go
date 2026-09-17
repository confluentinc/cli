package test

// TestQuery covers `confluent flink query` end to end against the mock Flink gateway.
// Scenarios are keyed off the `--sql` text; see buildQueryTestFixture in
// test/test-server/flink_gateway_router.go for what each one returns.
func (s *CLITestSuite) TestQuery() {
	tests := []CLITest{
		{args: "flink query --help", fixture: "query/help.golden"},

		{args: `flink query --sql "SELECT order_id, status FROM orders LIMIT 2;" --compute-pool lfcp-123456 --service-account sa-123456 --database lkc-123456`, fixture: "query/select.golden"},

		// Multi-page result set.
		{args: `flink query --sql "SELECT id FROM multi_page_table;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/multi-page.golden"},

		// -o json / -o yaml default to the schema+rows envelope. No statement name in
		// it (dropped to match the PRD's engine-agnostic envelope shape) and nothing
		// else random, so these are exact matches.
		{args: `flink query --sql "SELECT order_id, status FROM orders LIMIT 2;" --compute-pool lfcp-123456 --service-account sa-123456 -o json`, fixture: "query/select-json.golden"},
		{args: `flink query --sql "SELECT order_id, status FROM orders LIMIT 2;" --compute-pool lfcp-123456 --service-account sa-123456 -o yaml`, fixture: "query/select-yaml.golden"},

		// --raw drops the envelope, so this one is exact too.
		{args: `flink query --sql "SELECT order_id, status FROM orders LIMIT 2;" --compute-pool lfcp-123456 --service-account sa-123456 -o json --raw`, fixture: "query/select-raw.golden"},

		// --raw is meaningless for the default table output; rejected before a
		// statement is ever created, so no random name in the output.
		{args: `flink query --sql "SELECT order_id, status FROM orders LIMIT 2;" --compute-pool lfcp-123456 --service-account sa-123456 --raw`, fixture: "query/raw-without-serialized-output.golden", exitCode: 1},

		// --max-rows stops the drain early. Truncated is one of the two conditions that
		// makes runQuery's deferred cleanup stop the statement, so the name shows up again
		// in the "Stopped statement" message.
		{args: `flink query --sql "SELECT id FROM many_rows;" --compute-pool lfcp-123456 --service-account sa-123456 --max-rows 2`, fixture: "query/max-rows.golden", regex: true},

		// Regression case for a real false positive found against staging: a
		// LIMIT-satisfied read over a streaming source delivers every row (no next
		// token) but the job's phase stays RUNNING. No "may be incomplete" warning —
		// all rows print — but the deferred cleanup still stops the non-terminal
		// statement, same as it does after --max-rows.
		{args: `flink query --sql "SELECT id FROM limit_bounded_stream;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/limit-bounded-stream.golden", regex: true},

		// Non-append-only: an Operation column and a changelog warning, no stop (the
		// statement already reached a terminal phase on its own).
		{args: `flink query --sql "SELECT * FROM changelog;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/changelog.golden"},

		// RequireBounded rejects an unbounded statement before ever touching results.
		{args: `flink query --sql "SELECT * FROM unbounded_stream;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/unbounded.golden", regex: true, exitCode: 1},

		// The statement itself fails server-side.
		{args: `flink query --sql "SELECT * FROM will_fail;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/failed.golden", regex: true, exitCode: 1},

		// A results-conversion error (not Unbounded, not Canceled) falls through
		// handleQueryError's generic branch, which leaves settled false — the
		// deferred cleanup must still announce stopping the still-RUNNING
		// statement instead of only logging it.
		{args: `flink query --sql "SELECT * FROM unrecognized_column_type;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/unrecognized-column-type.golden", regex: true, exitCode: 1},

		// DDL has no result schema, so Run() returns before ever calling GetStatementResults.
		{args: `flink query --sql "CREATE TABLE t (id INT);" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/no-rows.golden", regex: true},

		// sql.snapshot.mode can't be overridden; this fails in buildQueryProperties, before
		// a statement is ever created, so there's no random name in the output.
		{args: `flink query --sql "SELECT 1;" --compute-pool lfcp-123456 --service-account sa-123456 --property sql.snapshot.mode=at-earliest`, fixture: "query/snapshot-mode-override.golden", exitCode: 1},

		// -f/--file reads the same SQL as the happy path, so it produces the same table.
		{args: "flink query -f test/fixtures/input/query/select.sql --compute-pool lfcp-123456 --service-account sa-123456 --database lkc-123456", fixture: "query/select.golden"},

		// --catalog is an alias for --environment and doesn't change what's printed.
		{args: `flink query --sql "SELECT order_id, status FROM orders LIMIT 2;" --compute-pool lfcp-123456 --service-account sa-123456 --catalog env-596 --database lkc-123456`, fixture: "query/select.golden"},

		// Cobra rejects the command before RunE runs unless exactly one of --sql/--file is set.
		{args: "flink query --compute-pool lfcp-123456 --service-account sa-123456", fixture: "query/missing-sql.golden", exitCode: 1},

		{args: `flink query --sql "SELECT 1;" --file test/fixtures/input/query/select.sql --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/sql-and-file.golden", exitCode: 1},

		// --cluster no longer exists; only --database sets the default Kafka cluster.
		{args: `flink query --sql "SELECT 1;" --compute-pool lfcp-123456 --service-account sa-123456 --cluster lkc-123456`, fixture: "query/unknown-cluster-flag.golden", exitCode: 1},

		// The SQL is no longer accepted positionally.
		{args: `flink query "SELECT 1;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/unexpected-positional-arg.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
