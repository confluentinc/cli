package test

func (s *CLITestSuite) TestConnectorCatalog() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{Args: "connector-catalog --help", Fixture: "connector-catalog/connector-catalog-help.golden"},
		{Args: "connector-catalog describe GcsSink --cluster lkc-123 -o json", Fixture: "connector-catalog/connector-catalog-describe-json.golden"},
		{Args: "connector-catalog describe GcsSink --cluster lkc-123 -o yaml", Fixture: "connector-catalog/connector-catalog-describe-yaml.golden"},
		{Args: "connector-catalog describe GcsSink --cluster lkc-123", Fixture: "connector-catalog/connector-catalog-describe.golden"},
		{Args: "connector-catalog list --cluster lkc-123 -o json", Fixture: "connector-catalog/connector-catalog-list-json.golden"},
		{Args: "connector-catalog list --cluster lkc-123 -o yaml", Fixture: "connector-catalog/connector-catalog-list-yaml.golden"},
		{Args: "connector-catalog list --cluster lkc-123", Fixture: "connector-catalog/connector-catalog-list.golden"},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunCcloudTest(tt)
	}
}
