package test

func (s *CLITestSuite) TestKSQL() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{Args: "ksql --help", Fixture: "ksql/ksql-help.golden"},
		{Args: "ksql app --help", Fixture: "ksql/ksql-app-help.golden"},
		{Args: "ksql app configure-acls --help", Fixture: "ksql/ksql-app-configure-acls-help.golden"},
		{Args: "ksql app create --help", Fixture: "ksql/ksql-app-create-help.golden"},
		{Args: "ksql app create test_ksql --cluster lkc-12345", Fixture: "ksql/ksql-app-create-result-deprecated.golden"},
		{Args: "ksql app create test_ksql_json --cluster lkc-12345 -o json", Fixture: "ksql/ksql-app-create-result-json-deprecated.golden"},
		{Args: "ksql app create test_ksql_yaml --cluster lkc-12345 -o yaml", Fixture: "ksql/ksql-app-create-result-yaml-deprecated.golden"},
		{Args: "ksql app create test_ksql --cluster lkc-12345 --api-key key --api-secret secret", Fixture: "ksql/ksql-app-create-result.golden"},
		{Args: "ksql app create test_ksql_json --cluster lkc-12345 --api-key key --api-secret secret -o json", Fixture: "ksql/ksql-app-create-result-json.golden"},
		{Args: "ksql app create test_ksql_yaml --cluster lkc-12345 --api-key key --api-secret secret -o yaml", Fixture: "ksql/ksql-app-create-result-yaml.golden"},
		{Args: "ksql app delete --help", Fixture: "ksql/ksql-app-delete-help.golden"},
		{Args: "ksql app delete lksqlc-12345", Fixture: "ksql/ksql-app-delete-result.golden"},
		{Args: "ksql app describe --help", Fixture: "ksql/ksql-app-describe-help.golden"},
		{Args: "ksql app describe lksqlc-12345 -o json", Fixture: "ksql/ksql-app-describe-result-json.golden"},
		{Args: "ksql app describe lksqlc-12345 -o yaml", Fixture: "ksql/ksql-app-describe-result-yaml.golden"},
		{Args: "ksql app describe lksqlc-12345", Fixture: "ksql/ksql-app-describe-result.golden"},
		{Args: "ksql app list --help", Fixture: "ksql/ksql-app-list-help.golden"},
		{Args: "ksql app list -o json", Fixture: "ksql/ksql-app-list-result-json.golden"},
		{Args: "ksql app list -o yaml", Fixture: "ksql/ksql-app-list-result-yaml.golden"},
		{Args: "ksql app list", Fixture: "ksql/ksql-app-list-result.golden"},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunCcloudTest(tt)
	}
}
