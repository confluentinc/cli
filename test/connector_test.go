package test

func (s *CLITestSuite) TestConnector() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "connector --help", fixture: "connector/connector-help.golden"},
		{args: "connector create --cluster lkc-123 --config test/fixtures/input/connector-config.yaml -o json", fixture: "connector/connector-create-json.golden"},
		{args: "connector create --cluster lkc-123 --config test/fixtures/input/connector-config.yaml -o yaml", fixture: "connector/connector-create-yaml.golden"},
		{args: "connector create --cluster lkc-123 --config test/fixtures/input/connector-config.yaml", fixture: "connector/connector-create.golden"},
		{args: "connector delete lcc-123 --cluster lkc-123", fixture: "connector/connector-delete.golden"},
		{args: "connector describe lcc-123 --cluster lkc-123 -o json", fixture: "connector/connector-describe-json.golden"},
		{args: "connector describe lcc-123 --cluster lkc-123 -o yaml", fixture: "connector/connector-describe-yaml.golden"},
		{args: "connector describe lcc-123 --cluster lkc-123", fixture: "connector/connector-describe.golden"},
		{args: "connector list --cluster lkc-123 -o json", fixture: "connector/connector-list-json.golden"},
		{args: "connector list --cluster lkc-123 -o yaml", fixture: "connector/connector-list-yaml.golden"},
		{args: "connector list --cluster lkc-123", fixture: "connector/connector-list.golden"},
		{args: "connector pause lcc-123 --cluster lkc-123", fixture: "connector/connector-pause.golden"},
		{args: "connector resume lcc-123 --cluster lkc-123", fixture: "connector/connector-resume.golden"},
		{args: "connector update lcc-123 --cluster lkc-123 --config test/fixtures/input/connector-config.yaml", fixture: "connector/connector-update.golden"},
		{args: "connector event describe", fixture: "connector/connector-event-describe.golden"},

		//Tests based on new config
		{args: "connector create --cluster lkc-123 --config test/fixtures/input/connector-config-new-format.json -o json", fixture: "connector/connector-create-new-config-json.golden"},
		{args: "connector create --cluster lkc-123 --config test/fixtures/input/connector-config-new-format.json -o yaml", fixture: "connector/connector-create-yaml.golden"},
		{args: "connector update lcc-123 --cluster lkc-123 --config test/fixtures/input/connector-config-new-format.json", fixture: "connector/connector-update.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCcloudTest(tt)
	}
}

func (s *CLITestSuite) TestConnectorCatalog() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "connector catalog --help", fixture: "connector/connector-catalog-help.golden"},
		{args: "connector catalog describe GcsSink --cluster lkc-123 -o json", fixture: "connector/connector-catalog-describe-json.golden"},
		{args: "connector catalog describe GcsSink --cluster lkc-123 -o yaml", fixture: "connector/connector-catalog-describe-yaml.golden"},
		{args: "connector catalog describe GcsSink --cluster lkc-123", fixture: "connector/connector-catalog-describe.golden"},
		{args: "connector catalog list --cluster lkc-123 -o json", fixture: "connector/connector-catalog-list-json.golden"},
		{args: "connector catalog list --cluster lkc-123 -o yaml", fixture: "connector/connector-catalog-list-yaml.golden"},
		{args: "connector catalog list --cluster lkc-123", fixture: "connector/connector-catalog-list.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCcloudTest(tt)
	}
}
