package test

func (s *CLITestSuite) TestConnect() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "connect --help", fixture: "connect/connect-help.golden"},
		{args: "connect create --cluster lkc-123 --config test/fixtures/input/connector-config.yaml -o json", fixture: "connect/connect-create-json.golden"},
		{args: "connect create --cluster lkc-123 --config test/fixtures/input/connector-config.yaml -o yaml", fixture: "connect/connect-create-yaml.golden"},
		{args: "connect create --cluster lkc-123 --config test/fixtures/input/connector-config.yaml", fixture: "connect/connect-create.golden"},
		{args: "connect delete lcc-123 --cluster lkc-123", fixture: "connect/connect-delete.golden"},
		{args: "connect describe lcc-123 --cluster lkc-123 -o json", fixture: "connect/connect-describe-json.golden"},
		{args: "connect describe lcc-123 --cluster lkc-123 -o yaml", fixture: "connect/connect-describe-yaml.golden"},
		{args: "connect describe lcc-123 --cluster lkc-123", fixture: "connect/connect-describe.golden"},
		{args: "connect list --cluster lkc-123 -o json", fixture: "connect/connect-list-json.golden"},
		{args: "connect list --cluster lkc-123 -o yaml", fixture: "connect/connect-list-yaml.golden"},
		{args: "connect list --cluster lkc-123", fixture: "connect/connect-list.golden"},
		{args: "connect pause lcc-123 --cluster lkc-123", fixture: "connect/connect-pause.golden"},
		{args: "connect resume lcc-123 --cluster lkc-123", fixture: "connect/connect-resume.golden"},
		{args: "connect update lcc-123 --cluster lkc-123 --config test/fixtures/input/connector-config.yaml", fixture: "connect/connect-update.golden"},
		{args: "connect event describe", fixture: "connect/connect-event-describe.golden"},

		//Tests based on new config
		{args: "connect create --cluster lkc-123 --config test/fixtures/input/connector-config-new-format.json -o json", fixture: "connect/connect-create-new-config-json.golden"},
		{args: "connect create --cluster lkc-123 --config test/fixtures/input/connector-config-new-format.json -o yaml", fixture: "connect/connect-create-yaml.golden"},
		{args: "connect update lcc-123 --cluster lkc-123 --config test/fixtures/input/connector-config-new-format.json", fixture: "connect/connect-update.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCloudTest(tt)
	}
}

func (s *CLITestSuite) TestConnectPlugin() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "connect plugin --help", fixture: "connect/connect-plugin-help.golden"},
		{args: "connect plugin describe GcsSink --cluster lkc-123 -o json", fixture: "connect/connect-plugin-describe-json.golden"},
		{args: "connect plugin describe GcsSink --cluster lkc-123 -o yaml", fixture: "connect/connect-plugin-describe-yaml.golden"},
		{args: "connect plugin describe GcsSink --cluster lkc-123", fixture: "connect/connect-plugin-describe.golden"},
		{args: "connect plugin list --cluster lkc-123 -o json", fixture: "connect/connect-plugin-list-json.golden"},
		{args: "connect plugin list --cluster lkc-123 -o yaml", fixture: "connect/connect-plugin-list-yaml.golden"},
		{args: "connect plugin list --cluster lkc-123", fixture: "connect/connect-plugin-list.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCloudTest(tt)
	}
}
