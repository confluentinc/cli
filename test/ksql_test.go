package test

func (s *CLITestSuite) TestKSQL() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "ksql --help", fixture: "ksql/help.golden"},
	}

	clusterTests := []CLITest{
		{args: "ksql cluster --help", fixture: "ksql/cluster/help.golden"},
		{args: "ksql cluster create --help", fixture: "ksql/cluster/create-help.golden"},
		{args: "ksql cluster create test_ksql --cluster lkc-12345", fixture: "ksql/cluster/create-result-missing-credential-identity.golden", wantErrCode: 1},
		{args: "ksql cluster create test_ksql --cluster lkc-12345 --credential-identity sa-credential-identity", fixture: "ksql/cluster/create-result.golden"},
		{args: "ksql cluster create test_ksql_json --cluster lkc-12345 --credential-identity sa-credential-identity -o json", fixture: "ksql/cluster/create-result-json.golden"},
		{args: "ksql cluster create test_ksql_yaml --cluster lkc-12345 --credential-identity sa-credential-identity -o yaml", fixture: "ksql/cluster/create-result-yaml.golden"},
		{args: "ksql cluster create test_ksql --cluster lkc-processLogFalse --credential-identity sa-credential-identity --log-exclude-rows", fixture: "ksql/cluster/create-result-log-exclude-rows.golden"},
		{args: "ksql cluster create test_ksql_json --cluster lkc-processLogFalse --credential-identity sa-credential-identity --log-exclude-rows -o json", fixture: "ksql/cluster/create-result-json-log-exclude-rows.golden"},
		{args: "ksql cluster create test_ksql_yaml --cluster lkc-processLogFalse --credential-identity sa-credential-identity --log-exclude-rows -o yaml", fixture: "ksql/cluster/create-result-yaml-log-exclude-rows.golden"},
		{args: "ksql cluster delete --help", fixture: "ksql/cluster/delete-help.golden"},
		{args: "ksql cluster delete lksqlc-12345 --force", fixture: "ksql/cluster/delete-result.golden"},
		{args: "ksql cluster delete lksqlc-12345", input: "account ksql\n", fixture: "ksql/cluster/delete-result-prompt.golden"},
		{args: "ksql cluster describe --help", fixture: "ksql/cluster/describe-help.golden"},
		{args: "ksql cluster describe lksqlc-12345 -o json", fixture: "ksql/cluster/describe-result-json.golden"},
		{args: "ksql cluster describe lksqlc-12345 -o yaml", fixture: "ksql/cluster/describe-result-yaml.golden"},
		{args: "ksql cluster describe lksqlc-12345", fixture: "ksql/cluster/describe-result.golden"},
		{args: "ksql cluster list --help", fixture: "ksql/cluster/list-help.golden"},
		{args: "ksql cluster list -o json", fixture: "ksql/cluster/list-result-json.golden"},
		{args: "ksql cluster list -o yaml", fixture: "ksql/cluster/list-result-yaml.golden"},
		{args: "ksql cluster list", fixture: "ksql/cluster/list-result.golden"},
	}

	tests = append(tests, clusterTests...)
	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestKsqlClusterConfigureAcls() {
	tests := []CLITest{
		{args: "ksql cluster configure-acls --help", fixture: "ksql/cluster/configure-acls-help.golden"},
		{args: "ksql cluster configure-acls lksqlc-12345 --cluster lkc-abcde", fixture: "ksql/cluster/configure-acls.golden"},
		{args: "ksql cluster configure-acls lksqlc-12345 --cluster lkc-abcde --dry-run", fixture: "ksql/cluster/configure-acls-dry-run.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
