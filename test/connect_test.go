package test

func (s *CLITestSuite) TestConnect() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "connect --help", fixture: "connect/help.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config.yaml -o json", fixture: "connect/cluster/create-json.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config.yaml -o yaml", fixture: "connect/cluster/create-yaml.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config.yaml", fixture: "connect/cluster/create.golden"},
		{args: "connect cluster delete lcc-123 --cluster lkc-123 --force", fixture: "connect/cluster/delete.golden"},
		{args: "connect cluster delete lcc-123 --cluster lkc-123", input: "az-connector\n", fixture: "connect/cluster/delete-prompt.golden"},
		{args: `__complete connect cluster describe ""`, useKafka: "lkc-123", fixture: "connect/cluster/describe-autocomplete.golden"},
		{args: "connect cluster describe lcc-123 --cluster lkc-123 -o json", fixture: "connect/cluster/describe-json.golden"},
		{args: "connect cluster describe lcc-123 --cluster lkc-123 -o yaml", fixture: "connect/cluster/describe-yaml.golden"},
		{args: "connect cluster describe lcc-123 --cluster lkc-123", fixture: "connect/cluster/describe.golden"},
		{args: "connect cluster list --cluster lkc-123 -o json", fixture: "connect/cluster/list-json.golden"},
		{args: "connect cluster list --cluster lkc-123 -o yaml", fixture: "connect/cluster/list-yaml.golden"},
		{args: "connect cluster list --cluster lkc-123", fixture: "connect/cluster/list.golden"},
		{args: "connect cluster update lcc-123 --cluster lkc-123 --config-file test/fixtures/input/connect/config.yaml", fixture: "connect/cluster/update.golden"},
		{args: "connect event describe", fixture: "connect/event-describe.golden"},

		// Tests based on new config
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config-new-format.json -o json", fixture: "connect/cluster/create-new-config-json.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config-new-format.json -o yaml", fixture: "connect/cluster/create-yaml.golden"},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config-malformed-new.json", fixture: "connect/cluster/create-malformed-new.golden", exitCode: 1},
		{args: "connect cluster create --cluster lkc-123 --config-file test/fixtures/input/connect/config-malformed-old.json", fixture: "connect/cluster/create-malformed-old.golden", exitCode: 1},
		{args: "connect cluster update lcc-123 --cluster lkc-123 --config-file test/fixtures/input/connect/config-new-format.json", fixture: "connect/cluster/update.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestConnectClusterPause() {
	tests := []CLITest{
		{args: "connect cluster pause --help", fixture: "connect/cluster/pause-help.golden"},
		{args: "connect cluster pause lcc-000000 --cluster lkc-123456", fixture: "connect/cluster/pause-unknown.golden", exitCode: 1},
		{args: "connect cluster pause lcc-123 --cluster lkc-123456", fixture: "connect/cluster/pause.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestConnectClusterResume() {
	tests := []CLITest{
		{args: "connect cluster resume --help", fixture: "connect/cluster/resume-help.golden"},
		{args: "connect cluster resume lcc-000000 --cluster lkc-123456", fixture: "connect/cluster/resume-unknown.golden", exitCode: 1},
		{args: "connect cluster resume lcc-123 --cluster lkc-123456", fixture: "connect/cluster/resume.golden"},
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
		{args: `__complete connect plugin describe ""`, useKafka: "lkc-123", fixture: "connect/plugin/describe-autocomplete.golden"},
		{args: "connect plugin describe GcsSink --cluster lkc-123 -o json", fixture: "connect/plugin/describe-json.golden"},
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
