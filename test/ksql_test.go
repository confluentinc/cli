package test

func (s *CLITestSuite) TestKSQL() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "ksql --help", fixture: "ksql/help.golden"},
	}

	appTests := []CLITest{
		{args: "ksql app --help", fixture: "ksql/app-help.golden"},
		{args: "ksql app configure-acls --help", fixture: "ksql/app-configure-acls-help.golden"},
		{args: "ksql app create --help", fixture: "ksql/app-create-help.golden"},
		{args: "ksql app create test_ksql --cluster lkc-12345", fixture: "ksql/app-create-result-missing-api-key.golden", wantErrCode: 1},
		{args: "ksql app create test_ksql --cluster lkc-12345 --api-key key --api-secret secret", fixture: "ksql/app-create-result.golden"},
		{args: "ksql app create test_ksql_json --cluster lkc-12345 --api-key key --api-secret secret -o json", fixture: "ksql/app-create-result-json.golden"},
		{args: "ksql app create test_ksql_yaml --cluster lkc-12345 --api-key key --api-secret secret -o yaml", fixture: "ksql/app-create-result-yaml.golden"},
		{args: "ksql app delete --help", fixture: "ksql/app-delete-help.golden"},
		{args: "ksql app delete lksqlc-12345", fixture: "ksql/app-delete-result.golden"},
		{args: "ksql app describe --help", fixture: "ksql/app-describe-help.golden"},
		{args: "ksql app describe lksqlc-12345 -o json", fixture: "ksql/app-describe-result-json.golden"},
		{args: "ksql app describe lksqlc-12345 -o yaml", fixture: "ksql/app-describe-result-yaml.golden"},
		{args: "ksql app describe lksqlc-12345", fixture: "ksql/app-describe-result.golden"},
		{args: "ksql app list --help", fixture: "ksql/app-list-help.golden"},
		{args: "ksql app list -o json", fixture: "ksql/app-list-result-json.golden"},
		{args: "ksql app list -o yaml", fixture: "ksql/app-list-result-yaml.golden"},
		{args: "ksql app list", fixture: "ksql/app-list-result.golden"},
	}

	clusterTests := []CLITest{
		{args: "ksql cluster --help", fixture: "ksql/cluster-help.golden"},
		{args: "ksql cluster configure-acls --help", fixture: "ksql/cluster-configure-acls-help.golden"},
		{args: "ksql cluster create --help", fixture: "ksql/cluster-create-help.golden"},
		{args: "ksql cluster create test_ksql --cluster lkc-12345", fixture: "ksql/cluster-create-result-missing-api-key.golden", wantErrCode: 1},
		{args: "ksql cluster create test_ksql --cluster lkc-12345 --api-key key --api-secret secret", fixture: "ksql/cluster-create-result.golden"},
		{args: "ksql cluster create test_ksql_json --cluster lkc-12345 --api-key key --api-secret secret -o json", fixture: "ksql/cluster-create-result-json.golden"},
		{args: "ksql cluster create test_ksql_yaml --cluster lkc-12345 --api-key key --api-secret secret -o yaml", fixture: "ksql/cluster-create-result-yaml.golden"},
		{args: "ksql cluster delete --help", fixture: "ksql/cluster-delete-help.golden"},
		{args: "ksql cluster delete lksqlc-12345", fixture: "ksql/cluster-delete-result.golden"},
		{args: "ksql cluster describe --help", fixture: "ksql/cluster-describe-help.golden"},
		{args: "ksql cluster describe lksqlc-12345 -o json", fixture: "ksql/cluster-describe-result-json.golden"},
		{args: "ksql cluster describe lksqlc-12345 -o yaml", fixture: "ksql/cluster-describe-result-yaml.golden"},
		{args: "ksql cluster describe lksqlc-12345", fixture: "ksql/cluster-describe-result.golden"},
		{args: "ksql cluster list --help", fixture: "ksql/cluster-list-help.golden"},
		{args: "ksql cluster list -o json", fixture: "ksql/cluster-list-result-json.golden"},
		{args: "ksql cluster list -o yaml", fixture: "ksql/cluster-list-result-yaml.golden"},
		{args: "ksql cluster list", fixture: "ksql/cluster-list-result.golden"},
	}

	tests = append(tests, appTests...)
	tests = append(tests, clusterTests...)
	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
