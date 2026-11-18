package test

// TestLightningQuery runs against the mock in test/test-server/lightning_router.go, which responds based on the SQL text.
func (s *CLITestSuite) TestLightningQuery() {
	tests := []CLITest{
		{args: `lightning query --sql "SELECT order_id, status FROM orders LIMIT 2" --cluster lkc-123456`, fixture: "lightning/query/select.golden"},
		{args: `lightning query --sql "SELECT order_id, status FROM orders LIMIT 2" --cluster lkc-123456 -o json`, fixture: "lightning/query/select-json.golden"},
		{args: `lightning query --sql "SELECT order_id, status FROM orders LIMIT 2" --cluster lkc-123456 -o yaml`, fixture: "lightning/query/select-yaml.golden"},

		// Columns keep the SELECT order, not alphabetical.
		{args: `lightning query --sql "SELECT status, order_id FROM orders LIMIT 2" --cluster lkc-123456 -o json`, fixture: "lightning/query/select-order-json.golden"},

		{args: "lightning query -f test/fixtures/input/lightning/query/select.sql --cluster lkc-123456", fixture: "lightning/query/select.golden"},
		// A byte order mark at the start of the file is dropped.
		{args: "lightning query -f test/fixtures/input/lightning/query/select-bom.sql --cluster lkc-123456", fixture: "lightning/query/select.golden"},

		// The mock echoes the environment and cluster the query is sent to.
		{args: `lightning query --sql "SELECT catalog_name, database_name FROM whoami"`, useKafka: "lkc-123456", fixture: "lightning/query/whoami-active-cluster.golden"},
		{args: `lightning query --sql "SELECT catalog_name, database_name FROM whoami" --environment env-123456 --cluster lkc-654321`, fixture: "lightning/query/whoami-flags.golden"},
		{args: `lightning query --sql "SELECT catalog_name, database_name FROM whoami" --context login-fake@user.com-http://127.0.0.1:1024`, useKafka: "lkc-123456", fixture: "lightning/query/whoami-active-cluster.golden"},

		// Values render as in `confluent flink query`.
		{args: `lightning query --sql "SELECT * FROM scalar_types" --cluster lkc-123456`, fixture: "lightning/query/scalar-types.golden"},
		{args: `lightning query --sql "SELECT * FROM scalar_types" --cluster lkc-123456 -o json`, fixture: "lightning/query/scalar-types-json.golden"},
		{args: `lightning query --sql "SELECT * FROM complex_types" --cluster lkc-123456`, fixture: "lightning/query/complex-types.golden"},
		{args: `lightning query --sql "SELECT * FROM complex_types" --cluster lkc-123456 -o json`, fixture: "lightning/query/complex-types-json.golden"},
		{args: `lightning query --sql "SELECT * FROM complex_types" --cluster lkc-123456 -o yaml`, fixture: "lightning/query/complex-types-yaml.golden"},

		// Control characters are escaped in the table.
		{args: `lightning query --sql "SELECT note FROM escapes" --cluster lkc-123456`, fixture: "lightning/query/escapes.golden"},

		// The row limit sent is the cap plus one, or none without a cap.
		{args: `lightning query --sql "SELECT max_result_rows FROM request_options" --cluster lkc-123456`, fixture: "lightning/query/request-options-human-default.golden"},
		{args: `lightning query --sql "SELECT max_result_rows FROM request_options" --cluster lkc-123456 --max-rows 1000`, fixture: "lightning/query/request-options-max-rows.golden"},
		{args: `lightning query --sql "SELECT max_result_rows FROM request_options" --cluster lkc-123456 -o json`, fixture: "lightning/query/request-options-json.golden"},

		// -o human shows 100 rows by default, with a notice after the table.
		{args: `lightning query --sql "SELECT id FROM hundreds" --cluster lkc-123456`, fixture: "lightning/query/human-cap.golden"},
		{args: `lightning query --sql "SELECT id FROM exactly_hundred" --cluster lkc-123456`, fixture: "lightning/query/exactly-hundred.golden"},
		{args: `lightning query --sql "SELECT id FROM hundreds" --cluster lkc-123456 --max-rows 0`, fixture: "lightning/query/human-uncap.golden"},
		{args: `lightning query --sql "SELECT id FROM hundreds" --cluster lkc-123456 -o json`, fixture: "lightning/query/hundreds-json.golden"},

		// --max-rows warns before the rows in every format.
		{args: `lightning query --sql "SELECT id FROM hundreds" --cluster lkc-123456 --max-rows 2`, fixture: "lightning/query/max-rows.golden"},
		{args: `lightning query --sql "SELECT id FROM hundreds" --cluster lkc-123456 --max-rows 2 -o json`, fixture: "lightning/query/max-rows-json.golden"},

		// No rows: a notice for -o human, an empty array for -o json/-o yaml.
		{args: `lightning query --sql "SELECT id FROM empty" --cluster lkc-123456`, fixture: "lightning/query/empty.golden"},
		{args: `lightning query --sql "SELECT id FROM empty" --cluster lkc-123456 -o json`, fixture: "lightning/query/empty-json.golden"},
		{args: `lightning query --sql "SELECT id FROM empty" --cluster lkc-123456 -o yaml`, fixture: "lightning/query/empty-yaml.golden"},

		// A repeated column name fails for -o json/-o yaml, whose keys must be unique.
		{args: `lightning query --sql "SELECT id, id FROM repeated_columns" --cluster lkc-123456`, fixture: "lightning/query/repeated-columns.golden"},
		{args: `lightning query --sql "SELECT id, id FROM repeated_columns" --cluster lkc-123456 -o json`, fixture: "lightning/query/repeated-columns-serialized.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT id, id FROM repeated_columns" --cluster lkc-123456 -o yaml`, fixture: "lightning/query/repeated-columns-serialized.golden", exitCode: 1},

		// A query that needs async execution fails, whether the server returns 422 or 202.
		{args: `lightning query --sql "SELECT * FROM huge" --cluster lkc-123456`, fixture: "lightning/query/async-required.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT * FROM promoted" --cluster lkc-123456`, fixture: "lightning/query/async-required.golden", exitCode: 1},

		// A gateway timeout suggests narrowing the query, without a request ID.
		{args: `lightning query --sql "SELECT * FROM slow" --cluster lkc-123456`, fixture: "lightning/query/gateway-timeout.golden", exitCode: 1},

		// Server errors show the server's message, plus the request ID for a 5xx.
		{args: `lightning query --sql "SELECT * FROM missing" --cluster lkc-123456`, fixture: "lightning/query/sql-error.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT * FROM server_error" --cluster lkc-123456`, fixture: "lightning/query/server-error.golden", exitCode: 1},

		// A result the CLI can't read fails instead of printing NULLs.
		{args: `lightning query --sql "SELECT id FROM unrecognized_column_type" --cluster lkc-123456`, fixture: "lightning/query/unrecognized-column-type.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT id FROM mismatched_row" --cluster lkc-123456`, fixture: "lightning/query/mismatched-row.golden", exitCode: 1},

		// Failures before the query is sent.
		{args: `lightning query --sql "SELECT 1"`, fixture: "lightning/query/no-cluster.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT 1" --cluster lkc-unknown`, fixture: "lightning/query/unknown-cluster.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT 1" --cluster lkc-not-found`, fixture: "lightning/query/cluster-not-found.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT 1" --cluster lkc-123456 --max-rows -1`, fixture: "lightning/query/negative-max-rows.golden", exitCode: 1},
		{args: `lightning query --sql "" --cluster lkc-123456`, fixture: "lightning/query/empty-sql.golden", exitCode: 1},

		// Exactly one of --sql and --file is required, and SQL isn't accepted as an argument.
		{args: "lightning query --cluster lkc-123456", fixture: "lightning/query/missing-sql.golden", exitCode: 1},
		{args: `lightning query --sql "SELECT 1" --file test/fixtures/input/lightning/query/select.sql --cluster lkc-123456`, fixture: "lightning/query/sql-and-file.golden", exitCode: 1},
		{args: `lightning query "SELECT 1" --cluster lkc-123456`, fixture: "lightning/query/unexpected-positional-arg.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}

	// Without a login, `lightning` shows the login error.
	s.runIntegrationTest(CLITest{args: "lightning", fixture: "lightning/not-logged-in.golden", exitCode: 1})
}
