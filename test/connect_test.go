package test

func (s *CLITestSuite) TestConnect() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "connect --help", fixture: "connect/help.golden"},
		{args: "connect create --cluster lkc-123 --config test/fixtures/input/connect/config.yaml -o json", fixture: "connect/create-json.golden"},
		{args: "connect create --cluster lkc-123 --config test/fixtures/input/connect/config.yaml -o yaml", fixture: "connect/create-yaml.golden"},
		{args: "connect create --cluster lkc-123 --config test/fixtures/input/connect/config.yaml", fixture: "connect/create.golden"},
		{args: "connect delete lcc-123 --cluster lkc-123", fixture: "connect/delete.golden"},
		{args: "connect describe lcc-123 --cluster lkc-123 -o json", fixture: "connect/describe-json.golden"},
		{args: "connect describe lcc-123 --cluster lkc-123 -o yaml", fixture: "connect/describe-yaml.golden"},
		{args: "connect describe lcc-123 --cluster lkc-123", fixture: "connect/describe.golden"},
		{args: "connect list --cluster lkc-123 -o json", fixture: "connect/list-json.golden"},
		{args: "connect list --cluster lkc-123 -o yaml", fixture: "connect/list-yaml.golden"},
		{args: "connect list --cluster lkc-123", fixture: "connect/list.golden"},
		{args: "connect update lcc-123 --cluster lkc-123 --config test/fixtures/input/connect/config.yaml", fixture: "connect/update.golden"},
		{args: "connect event describe", fixture: "connect/event-describe.golden"},

		//Tests based on new config
		{args: "connect create --cluster lkc-123 --config test/fixtures/input/connect/config-new-format.json -o json", fixture: "connect/create-new-config-json.golden"},
		{args: "connect create --cluster lkc-123 --config test/fixtures/input/connect/config-new-format.json -o yaml", fixture: "connect/create-yaml.golden"},
		{args: "connect update lcc-123 --cluster lkc-123 --config test/fixtures/input/connect/config-new-format.json", fixture: "connect/update.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestConnectPause() {
	tests := []CLITest{
		{args: "connect pause --help", fixture: "connect/pause-help.golden"},
		{args: "connect pause lcc-000000 --cluster lkc-123456", fixture: "connect/pause-unknown.golden", wantErrCode: 1},
		{args: "connect pause lcc-123 --cluster lkc-123456", fixture: "connect/pause.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestConnectResume() {
	tests := []CLITest{
		{args: "connect resume --help", fixture: "connect/resume-help.golden"},
		{args: "connect resume lcc-000000 --cluster lkc-123456", fixture: "connect/resume-unknown.golden", wantErrCode: 1},
		{args: "connect resume lcc-123 --cluster lkc-123456", fixture: "connect/resume.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestConnectPlugin() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "connect plugin --help", fixture: "connect/plugin/help.golden"},
		{args: "connect plugin describe GcsSink --cluster lkc-123 -o json", fixture: "connect/plugin/describe-json.golden"},
		{args: "connect plugin describe GcsSink --cluster lkc-123 -o yaml", fixture: "connect/plugin/describe-yaml.golden"},
		{args: "connect plugin describe GcsSink --cluster lkc-123", fixture: "connect/plugin/describe.golden"},
		{args: "connect plugin list --cluster lkc-123 -o json", fixture: "connect/plugin/list-json.golden"},
		{args: "connect plugin list --cluster lkc-123 -o yaml", fixture: "connect/plugin/list-yaml.golden"},
		{args: "connect plugin list --cluster lkc-123", fixture: "connect/plugin/list.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
