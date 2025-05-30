package test

func (s *CLITestSuite) TestKsql() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "ksql cluster create test_ksql --cluster lkc-12345 --credential-identity sa-credential-identity", fixture: "ksql/cluster/create-result.golden"},
		{args: "ksql cluster create test_ksql --cluster lkc-12345", fixture: "ksql/cluster/create-result-missing-credential-identity.golden", exitCode: 1},
		{args: "ksql cluster create test_ksql --cluster lkc-processLogFalse --credential-identity sa-credential-identity --log-exclude-rows", fixture: "ksql/cluster/create-result-log-exclude-rows.golden"},
		{args: "ksql cluster create test_ksql_yaml --cluster lkc-12345 --credential-identity sa-credential-identity -o yaml", fixture: "ksql/cluster/create-result-yaml.golden"},
		{args: "ksql cluster create test_ksql_yaml --cluster lkc-processLogFalse --credential-identity sa-credential-identity --log-exclude-rows -o yaml", fixture: "ksql/cluster/create-result-yaml-log-exclude-rows.golden"},
		{args: "ksql cluster delete lksqlc-12345 --force", fixture: "ksql/cluster/delete-result.golden"},
		{args: "ksql cluster delete lksqlc-12345 lksqlc-12346", fixture: "ksql/cluster/delete-multiple-fail.golden", exitCode: 1},
		{args: "ksql cluster delete lksqlc-12345 lksqlc-ksql1", input: "n\n", fixture: "ksql/cluster/delete-multiple-refuse.golden"},
		{args: "ksql cluster delete lksqlc-12345 lksqlc-ksql1", input: "y\n", fixture: "ksql/cluster/delete-multiple-success.golden"},
		{args: "ksql cluster delete lksqlc-12345", input: "y\n", fixture: "ksql/cluster/delete-result-prompt.golden"},
		{args: "ksql cluster describe lksqlc-12345 -o yaml", fixture: "ksql/cluster/describe-result-yaml.golden"},
		{args: "ksql cluster describe lksqlc-12345", fixture: "ksql/cluster/describe-result.golden"},
		{args: "ksql cluster list -o yaml", fixture: "ksql/cluster/list-result-yaml.golden"},
		{args: "ksql cluster list", fixture: "ksql/cluster/list-result.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKsqlClusterConfigureAcls() {
	tests := []CLITest{
		{args: "ksql cluster configure-acls lksqlc-12345 --cluster lkc-abcde", fixture: "ksql/cluster/configure-acls.golden"},
		{args: "ksql cluster configure-acls lksqlc-12345 --cluster lkc-abcde --dry-run", fixture: "ksql/cluster/configure-acls-dry-run.golden"},
		{args: "ksql cluster configure-acls lksqlc-ksql1 --cluster lkc-12345", fixture: "ksql/cluster/configure-acls-no-service-account.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKsql_Autocomplete() {
	tests := []CLITest{
		{args: `__complete ksql cluster describe ""`, fixture: "ksql/cluster/describe-autocomplete.golden"},
		{args: `__complete ksql cluster create my-cluster --credential-identity ""`, fixture: "ksql/cluster/create-credential-identity-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
