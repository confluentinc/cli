package test

func (s *CLITestSuite) TestKSQL() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "ksql cluster create test_ksql --cluster lkc-12345", fixture: "ksql/cluster/create-result-missing-credential-identity.golden", exitCode: 1},
		{args: "ksql cluster create test_ksql --cluster lkc-12345 --credential-identity sa-credential-identity", fixture: "ksql/cluster/create-result.golden"},
		{args: "ksql cluster create test_ksql_json --cluster lkc-12345 --credential-identity sa-credential-identity -o json", fixture: "ksql/cluster/create-result-json.golden"},
		{args: "ksql cluster create test_ksql_yaml --cluster lkc-12345 --credential-identity sa-credential-identity -o yaml", fixture: "ksql/cluster/create-result-yaml.golden"},
		{args: "ksql cluster create test_ksql --cluster lkc-processLogFalse --credential-identity sa-credential-identity --log-exclude-rows", fixture: "ksql/cluster/create-result-log-exclude-rows.golden"},
		{args: "ksql cluster create test_ksql_json --cluster lkc-processLogFalse --credential-identity sa-credential-identity --log-exclude-rows -o json", fixture: "ksql/cluster/create-result-json-log-exclude-rows.golden"},
		{args: "ksql cluster create test_ksql_yaml --cluster lkc-processLogFalse --credential-identity sa-credential-identity --log-exclude-rows -o yaml", fixture: "ksql/cluster/create-result-yaml-log-exclude-rows.golden"},
		{args: "ksql cluster delete lksqlc-12345 --force", fixture: "ksql/cluster/delete-result.golden"},
		{args: "ksql cluster delete lksqlc-12345", input: "account ksql\n", fixture: "ksql/cluster/delete-result-prompt.golden"},
		{args: "ksql cluster describe lksqlc-12345 -o json", fixture: "ksql/cluster/describe-result-json.golden"},
		{args: "ksql cluster describe lksqlc-12345 -o yaml", fixture: "ksql/cluster/describe-result-yaml.golden"},
		{args: "ksql cluster describe lksqlc-12345", fixture: "ksql/cluster/describe-result.golden"},
		{args: "ksql cluster list -o json", fixture: "ksql/cluster/list-result-json.golden"},
		{args: "ksql cluster list -o yaml", fixture: "ksql/cluster/list-result-yaml.golden"},
		{args: "ksql cluster list", fixture: "ksql/cluster/list-result.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestKsqlClusterConfigureAcls() {
	tests := []CLITest{
		{args: "ksql cluster configure-acls lksqlc-12345 --cluster lkc-abcde", fixture: "ksql/cluster/configure-acls.golden"},
		{args: "ksql cluster configure-acls lksqlc-12345 --cluster lkc-abcde --dry-run", fixture: "ksql/cluster/configure-acls-dry-run.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestKSQLAutocomplete() {
	test := CLITest{args: `__complete ksql cluster describe ""`, login: "cloud", fixture: "ksql/cluster/describe-autocomplete.golden"}
	s.runIntegrationTest(test)
}
