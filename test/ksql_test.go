package test

func (s *CLITestSuite) TestKSQLCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		// Show what commands are available
		{args: "ksql --help", fixture: "ksql-help.golden"},
		{args: "ksql app --help", fixture: "ksql-app-help.golden"},
		{args: "ksql app configure-acls --help", fixture: "ksql-app-configure-acls-help.golden"},
		{args: "ksql app create --help", fixture: "ksql-app-create-help.golden"},
		{args: "ksql app delete --help", fixture: "ksql-app-delete-help.golden"},
		{args: "ksql app describe --help", fixture: "ksql-app-describe-help.golden"},
		{args: "ksql app list --help", fixture: "ksql-app-list-help.golden"},

		{args: "ksql app create test_ksql --storage 101 --cluster lkc-12345", fixture: "ksql-app-create-result.golden"},
		{args: "ksql app describe lksqlc-12345", fixture: "ksql-app-describe-result.golden"},
		{args: "ksql app describe lksqlc-12345 -o json", fixture: "ksql-app-describe-result-json.golden"},
		{args: "ksql app describe lksqlc-12345 -o yaml", fixture: "ksql-app-describe-result-yaml.golden"},
		{args: "ksql app list", fixture: "ksql-app-list-result.golden"},
		{args: "ksql app list -o json", fixture: "ksql-app-list-result-json.golden"},
		{args: "ksql app list -o yaml", fixture: "ksql-app-list-result-yaml.golden"},
		{args: "ksql app delete lksqlc-12345", fixture: "ksql-app-delete-result.golden"},
	}
	resetConfiguration(s.T(), "ccloud")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.login = "default"
		tt.workflow = true
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL, kafkaAPIURL)
	}
}
