package test

func (s *CLITestSuite) TestConnectorCatalog() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "connector-catalog --help", fixture: "connector-catalog/connector-catalog-help.golden"},
		{args: "connector-catalog describe GcsSink --cluster lkc-123 -o json", fixture: "connector-catalog/connector-catalog-describe-json.golden"},
		{args: "connector-catalog describe GcsSink --cluster lkc-123 -o yaml", fixture: "connector-catalog/connector-catalog-describe-yaml.golden"},
		{args: "connector-catalog describe GcsSink --cluster lkc-123", fixture: "connector-catalog/connector-catalog-describe.golden"},
		{args: "connector-catalog list --cluster lkc-123 -o json", fixture: "connector-catalog/connector-catalog-list-json.golden"},
		{args: "connector-catalog list --cluster lkc-123 -o yaml", fixture: "connector-catalog/connector-catalog-list-yaml.golden"},
		{args: "connector-catalog list --cluster lkc-123", fixture: "connector-catalog/connector-catalog-list.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCcloudTest(tt)
	}
}
