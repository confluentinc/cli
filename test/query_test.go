package test

// TestQuery covers `confluent query` end to end against the mock Flink gateway.
// Scenarios are keyed off the `--sql` text; see buildQueryTestFixture in
// test/test-server/flink_gateway_router.go for what each one returns.
func (s *CLITestSuite) TestQuery() {
	tests := []CLITest{
		{args: "query --help", fixture: "query/help.golden"},

		{args: `query --sql "SELECT order_id, status FROM orders LIMIT 2;" --compute-pool lfcp-123456 --service-account sa-123456 --database lkc-123456`, fixture: "query/select.golden"},

		// Positional SQL instead of --sql, and a multi-page result set.
		{args: `query "SELECT id FROM multi_page_table;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/multi-page.golden"},

		// -o json / -o yaml default to the schema+rows envelope; the statement name it
		// carries is random per run (types.GenerateStatementName), so these are regexes.
		{args: `query --sql "SELECT order_id, status FROM orders LIMIT 2;" --compute-pool lfcp-123456 --service-account sa-123456 -o json`, fixture: "query/select-json.golden", regex: true},
		{args: `query --sql "SELECT order_id, status FROM orders LIMIT 2;" --compute-pool lfcp-123456 --service-account sa-123456 -o yaml`, fixture: "query/select-yaml.golden", regex: true},

		// --raw drops the envelope (and the statement name with it), so this one is exact.
		{args: `query --sql "SELECT order_id, status FROM orders LIMIT 2;" --compute-pool lfcp-123456 --service-account sa-123456 -o json --raw`, fixture: "query/select-raw.golden"},

		// --max-rows stops the drain early. Truncated is one of the two conditions that
		// makes runQuery's deferred cleanup stop the statement, so the name shows up again
		// in the "Stopped statement" message.
		{args: `query --sql "SELECT id FROM many_rows;" --compute-pool lfcp-123456 --service-account sa-123456 --max-rows 2`, fixture: "query/max-rows.golden", regex: true},

		// Non-append-only: an Operation column and a changelog warning, no stop (the
		// statement already reached a terminal phase on its own).
		{args: `query --sql "SELECT * FROM changelog;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/changelog.golden"},

		// RequireBounded rejects an unbounded statement before ever touching results.
		{args: `query --sql "SELECT * FROM unbounded_stream;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/unbounded.golden", regex: true, exitCode: 1},

		// The statement itself fails server-side.
		{args: `query --sql "SELECT * FROM will_fail;" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/failed.golden", regex: true, exitCode: 1},

		// DDL has no result schema, so Run() returns before ever calling GetStatementResults.
		{args: `query --sql "CREATE TABLE t (id INT);" --compute-pool lfcp-123456 --service-account sa-123456`, fixture: "query/no-rows.golden", regex: true},

		// sql.snapshot.mode can't be overridden; this fails in buildQueryProperties, before
		// a statement is ever created, so there's no random name in the output.
		{args: `query --sql "SELECT 1;" --compute-pool lfcp-123456 --service-account sa-123456 --property sql.snapshot.mode=at-earliest`, fixture: "query/snapshot-mode-override.golden", exitCode: 1},

		// -f/--file reads the same SQL as the happy path, so it produces the same table.
		{args: "query -f test/fixtures/input/query/select.sql --compute-pool lfcp-123456 --service-account sa-123456 --database lkc-123456", fixture: "query/select.golden"},

		// --catalog/--cluster are aliases for --environment/--database and don't change
		// what's printed.
		{args: `query --sql "SELECT order_id, status FROM orders LIMIT 2;" --compute-pool lfcp-123456 --service-account sa-123456 --catalog env-596 --cluster lkc-123456`, fixture: "query/select.golden"},

		// resolveSQL requires exactly one of --sql, --file or the positional argument.
		{args: "query --compute-pool lfcp-123456 --service-account sa-123456", fixture: "query/missing-sql.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
