package test

// TestLightningQuery covers `confluent lightning query` end to end against the mock Lightning
// endpoint. Scenarios are keyed off the `--sql` text; see handleLightningQuery in
// test/test-server/lightning_router.go for what each one returns.
func (s *CLITestSuite) TestLightningQuery() {
	tests := []CLITest{
		// `lightning query --help` is covered by TestHelp, which walks the whole command tree.

		{args: `lightning query --sql "SELECT order_id, status FROM orders LIMIT 2" --cluster lkc-123456`, fixture: "lightning/query/select.golden"},
		{args: `lightning query --sql "SELECT order_id, status FROM orders LIMIT 2" --cluster lkc-123456 -o json`, fixture: "lightning/query/select-json.golden"},
		{args: `lightning query --sql "SELECT order_id, status FROM orders LIMIT 2" --cluster lkc-123456 -o yaml`, fixture: "lightning/query/select-yaml.golden"},

		// Columns keep the SELECT order (status before order_id), not alphabetical.
		{args: `lightning query --sql "SELECT status, order_id FROM orders LIMIT 2" --cluster lkc-123456 -o json`, fixture: "lightning/query/select-order-json.golden"},

		// -f/--file reads the same SQL as the happy path, so it produces the same table.
		{args: "lightning query -f test/fixtures/input/lightning/query/select.sql --cluster lkc-123456", fixture: "lightning/query/select.golden"},

		// The mock echoes the environment and cluster each query is sent to: the active ones by
		// default, or the ones passed with --environment and --cluster.
		{args: `lightning query --sql "SELECT catalog_name, database_name FROM whoami"`, useKafka: "lkc-123456", fixture: "lightning/query/whoami-active-cluster.golden"},
		{args: `lightning query --sql "SELECT catalog_name, database_name FROM whoami" --environment env-123456 --cluster lkc-654321`, fixture: "lightning/query/whoami-flags.golden"},
		// --context picks the context whose environment and Kafka cluster are used.
		{args: `lightning query --sql "SELECT catalog_name, database_name FROM whoami" --context login-fake@user.com-http://127.0.0.1:1024`, useKafka: "lkc-123456", fixture: "lightning/query/whoami-active-cluster.golden"},

		// Values render like `confluent flink query`: typed in -o json/-o yaml, text in the table.
		{args: `lightning query --sql "SELECT * FROM scalar_types" --cluster lkc-123456`, fixture: "lightning/query/scalar-types.golden"},
		{args: `lightning query --sql "SELECT * FROM scalar_types" --cluster lkc-123456 -o json`, fixture: "lightning/query/scalar-types-json.golden"},
		{args: `lightning query --sql "SELECT * FROM complex_types" --cluster lkc-123456`, fixture: "lightning/query/complex-types.golden"},
		{args: `lightning query --sql "SELECT * FROM complex_types" --cluster lkc-123456 -o json`, fixture: "lightning/query/complex-types-json.golden"},
		{args: `lightning query --sql "SELECT * FROM complex_types" --cluster lkc-123456 -o yaml`, fixture: "lightning/query/complex-types-yaml.golden"},

		// Control characters in a value are escaped in the table instead of reaching the terminal.
		{args: `lightning query --sql "SELECT note FROM escapes" --cluster lkc-123456`, fixture: "lightning/query/escapes.golden"},

		// The row limit sent with the query: the cap plus one extra row (to tell a cut result from one
		// that fits exactly) while that fits the server's limit of 200, and none otherwise.
		{args: `lightning query --sql "SELECT max_result_rows FROM request_options" --cluster lkc-123456`, fixture: "lightning/query/request-options-human-default.golden"},
		{args: `lightning query --sql "SELECT max_result_rows FROM request_options" --cluster lkc-123456 --max-rows 199`, fixture: "lightning/query/request-options-max-rows-199.golden"},
		{args: `lightning query --sql "SELECT max_result_rows FROM request_options" --cluster lkc-123456 --max-rows 200`, fixture: "lightning/query/request-options-max-rows-200.golden"},
		{args: `lightning query --sql "SELECT max_result_rows FROM request_options" --cluster lkc-123456 -o json`, fixture: "lightning/query/request-options-json.golden"},

		// -o human caps at 100 by default, with a notice after the table; the mock returns the
		// one extra row the CLI asks for, which is how it knows the result was cut.
		{args: `lightning query --sql "SELECT id FROM hundreds" --cluster lkc-123456`, fixture: "lightning/query/human-cap.golden"},
		// Exactly 100 rows under the default cap: not truncated, so no notice.
		{args: `lightning query --sql "SELECT id FROM exactly_hundred" --cluster lkc-123456`, fixture: "lightning/query/exactly-hundred.golden"},
		// An explicit --max-rows 0 opts -o human out of the default cap.
		{args: `lightning query --sql "SELECT id FROM hundreds" --cluster lkc-123456 --max-rows 0`, fixture: "lightning/query/human-uncap.golden"},
		// -o json stays uncapped.
		{args: `lightning query --sql "SELECT id FROM hundreds" --cluster lkc-123456 -o json`, fixture: "lightning/query/hundreds-json.golden"},

		// An explicit --max-rows warns before the rows, for every output format.
		{args: `lightning query --sql "SELECT id FROM hundreds" --cluster lkc-123456 --max-rows 2`, fixture: "lightning/query/max-rows.golden"},
		{args: `lightning query --sql "SELECT id FROM hundreds" --cluster lkc-123456 --max-rows 2 -o json`, fixture: "lightning/query/max-rows-json.golden"},
		// Without a row limit of its own, a query gets the server's limit of 200 rows, and the response
		// doesn't say whether that left rows out, so a result that size warns that it may be truncated.
		{args: `lightning query --sql "SELECT id FROM two_hundred_one" --cluster lkc-123456 --max-rows 200`, fixture: "lightning/query/server-limit.golden"},
		{args: `lightning query --sql "SELECT id FROM two_hundred_one" --cluster lkc-123456 -o yaml`, fixture: "lightning/query/server-limit-yaml.golden"},

		// A schema with zero rows: a notice for -o human, an empty array for -o json/-o yaml.
		{args: `lightning query --sql "SELECT id FROM empty" --cluster lkc-123456`, fixture: "lightning/query/empty.golden"},
		{args: `lightning query --sql "SELECT id FROM empty" --cluster lkc-123456 -o json`, fixture: "lightning/query/empty-json.golden"},
		{args: `lightning query --sql "SELECT id FROM empty" --cluster lkc-123456 -o yaml`, fixture: "lightning/query/empty-yaml.golden"},

		// A query that can't be answered synchronously fails with a suggestion, whether the server
		// rejects it (422) or, ignoring the sync-only header, accepts it as an asynchronous job (202).
		{args: `lightning query --sql "SELECT * FROM huge" --cluster lkc-123456`, fixture: "lightning/query/async-required.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT * FROM promoted" --cluster lkc-123456`, fixture: "lightning/query/async-required.golden", exitCode: 1},

		// Server-side failures surface the server's message, and server errors their request ID.
		{args: `lightning query --sql "SELECT * FROM missing" --cluster lkc-123456`, fixture: "lightning/query/sql-error.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT * FROM server_error" --cluster lkc-123456`, fixture: "lightning/query/server-error.golden", exitCode: 1},

		// A result the CLI can't read fails instead of printing NULLs.
		{args: `lightning query --sql "SELECT id FROM unrecognized_column_type" --cluster lkc-123456`, fixture: "lightning/query/unrecognized-column-type.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT id FROM mismatched_row" --cluster lkc-123456`, fixture: "lightning/query/mismatched-row.golden", exitCode: 1},

		// Failures before the query is sent.
		{args: `lightning query --sql "SELECT 1"`, fixture: "lightning/query/no-cluster.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT 1" --cluster lkc-unknown`, fixture: "lightning/query/unknown-cluster.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT 1" --cluster lkc-123456 --max-rows -1`, fixture: "lightning/query/negative-max-rows.golden", exitCode: 1},
		{args: `lightning query --sql "" --cluster lkc-123456`, fixture: "lightning/query/empty-sql.golden", exitCode: 1},

		// Cobra rejects the command before RunE runs unless exactly one of --sql/--file is set.
		{args: "lightning query --cluster lkc-123456", fixture: "lightning/query/missing-sql.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT 1" --file test/fixtures/input/lightning/query/select.sql --cluster lkc-123456`, fixture: "lightning/query/sql-and-file.golden", exitCode: 1},

		// The SQL is not accepted positionally.
		{args: `lightning query "SELECT 1" --cluster lkc-123456`, fixture: "lightning/query/unexpected-positional-arg.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}

	// Without a login, the `lightning` parent fails with the login error instead of an empty help page.
	s.runIntegrationTest(CLITest{args: "lightning", fixture: "lightning/not-logged-in.golden", exitCode: 1})
}
