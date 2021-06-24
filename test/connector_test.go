package test

func (s *CLITestSuite) TestConnector() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{Args: "connector --help", Fixture: "connector/connector-help.golden"},
		{Args: "connector create --cluster lkc-123 --config test/Fixtures/input/connector-config.yaml -o json", Fixture: "connector/connector-create-json.golden"},
		{Args: "connector create --cluster lkc-123 --config test/Fixtures/input/connector-config.yaml -o yaml", Fixture: "connector/connector-create-yaml.golden"},
		{Args: "connector create --cluster lkc-123 --config test/Fixtures/input/connector-config.yaml", Fixture: "connector/connector-create.golden"},
		{Args: "connector delete lcc-123 --cluster lkc-123", Fixture: "connector/connector-delete.golden"},
		{Args: "connector describe lcc-123 --cluster lkc-123 -o json", Fixture: "connector/connector-describe-json.golden"},
		{Args: "connector describe lcc-123 --cluster lkc-123 -o yaml", Fixture: "connector/connector-describe-yaml.golden"},
		{Args: "connector describe lcc-123 --cluster lkc-123", Fixture: "connector/connector-describe.golden"},
		{Args: "connector list --cluster lkc-123 -o json", Fixture: "connector/connector-list-json.golden"},
		{Args: "connector list --cluster lkc-123 -o yaml", Fixture: "connector/connector-list-yaml.golden"},
		{Args: "connector list --cluster lkc-123", Fixture: "connector/connector-list.golden"},
		{Args: "connector pause lcc-123 --cluster lkc-123", Fixture: "connector/connector-pause.golden"},
		{Args: "connector resume lcc-123 --cluster lkc-123", Fixture: "connector/connector-resume.golden"},
		{Args: "connector update lcc-123 --cluster lkc-123 --config test/Fixtures/input/connector-config.yaml", Fixture: "connector/connector-update.golden"},
		{Args: "connector event describe", Fixture: "connector/connector-event-describe.golden"},

		//Tests based on new config
		{Args: "connector create --cluster lkc-123 --config test/Fixtures/input/connector-config-new-format.json -o json", Fixture: "connector/connector-create-new-config-json.golden"},
		{Args: "connector create --cluster lkc-123 --config test/Fixtures/input/connector-config-new-format.json -o yaml", Fixture: "connector/connector-create-yaml.golden"},
		{Args: "connector update lcc-123 --cluster lkc-123 --config test/Fixtures/input/connector-config-new-format.json", Fixture: "connector/connector-update.golden"},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunCcloudTest(tt)
	}
}
