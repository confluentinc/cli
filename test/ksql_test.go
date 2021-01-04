package test

func (s *CLITestSuite) TestKSQL() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "ksql --help", fixture: "ksql/ksql-help.golden"},
		{args: "ksql app --help", fixture: "ksql/ksql-app-help.golden"},
		{args: "ksql app configure-acls --help", fixture: "ksql/ksql-app-configure-acls-help.golden"},
		{args: "ksql app create --help", fixture: "ksql/ksql-app-create-help.golden"},
		{args: "ksql app create test_ksql --cluster lkc-12345", fixture: "ksql/ksql-app-create-result.golden"},
		{args: "ksql app create test_ksql_json --cluster lkc-12345 -o json", fixture: "ksql/ksql-app-create-result-json.golden"},
		{args: "ksql app create test_ksql_yaml --cluster lkc-12345 -o yaml", fixture: "ksql/ksql-app-create-result-yaml.golden"},
		{args: "ksql app delete --help", fixture: "ksql/ksql-app-delete-help.golden"},
		{args: "ksql app delete lksqlc-12345", fixture: "ksql/ksql-app-delete-result.golden"},
		{args: "ksql app describe --help", fixture: "ksql/ksql-app-describe-help.golden"},
		{args: "ksql app describe lksqlc-12345 -o json", fixture: "ksql/ksql-app-describe-result-json.golden"},
		{args: "ksql app describe lksqlc-12345 -o yaml", fixture: "ksql/ksql-app-describe-result-yaml.golden"},
		{args: "ksql app describe lksqlc-12345", fixture: "ksql/ksql-app-describe-result.golden"},
		{args: "ksql app list --help", fixture: "ksql/ksql-app-list-help.golden"},
		{args: "ksql app list -o json", fixture: "ksql/ksql-app-list-result-json.golden"},
		{args: "ksql app list -o yaml", fixture: "ksql/ksql-app-list-result-yaml.golden"},
		{args: "ksql app list", fixture: "ksql/ksql-app-list-result.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCcloudTest(tt)
	}
}
