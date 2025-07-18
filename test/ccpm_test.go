package test

func (s *CLITestSuite) TestCCPM() {
	tests := []CLITest{
		{args: "ccpm --help", fixture: "ccpm/help.golden"},
		{args: "ccpm plugin --help", fixture: "ccpm/plugin-help.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestCCPMPlugin() {
	tests := []CLITest{
		// Plugin list tests
		{args: "ccpm plugin list --environment env-123456", fixture: "ccpm/plugin-list.golden"},
		{args: "ccpm plugin list --environment env-123456 --cloud aws", fixture: "ccpm/plugin-list-aws.golden"},
		{args: "ccpm plugin list --environment env-123456 -o json", fixture: "ccpm/plugin-list-json.golden"},
		{args: "ccpm plugin list --environment env-123456 -o yaml", fixture: "ccpm/plugin-list-yaml.golden"},

		// Plugin create tests
		{args: "ccpm plugin create --name my-custom-plugin --description Test_plugin --cloud aws --environment env-123456", fixture: "ccpm/plugin-create.golden"},
		{args: "ccpm plugin create --name my-custom-plugin --cloud aws --environment env-123456", fixture: "ccpm/plugin-create-without-description.golden"},
		{args: "ccpm plugin create --name my-custom-plugin --cloud invalid-cloud --environment env-123456", fixture: "ccpm/plugin-create-invalid-cloud.golden", exitCode: 1},

		// Plugin describe tests - using exact IDs from test-server
		{args: "ccpm plugin describe ccp-123456 --environment env-123456", fixture: "ccpm/plugin-describe.golden"},
		{args: "ccpm plugin describe ccp-123456 --environment env-123456 -o json", fixture: "ccpm/plugin-describe-json.golden"},
		{args: "ccpm plugin describe ccp-123456 --environment env-123456 -o yaml", fixture: "ccpm/plugin-describe-yaml.golden"},
		{args: "ccpm plugin describe invalid-id --environment env-123456", fixture: "ccpm/plugin-describe-not-found.golden", exitCode: 1},

		// Plugin update tests - using exact IDs from test-server
		{args: "ccpm plugin update ccp-123456 --name updated-plugin-name --environment env-123456", fixture: "ccpm/plugin-update.golden"},
		{args: "ccpm plugin update ccp-123456 --description 'Updated description' --environment env-123456", fixture: "ccpm/plugin-update.golden"},
		{args: "ccpm plugin update ccp-123456 --name updated-name --description 'Updated description' --environment env-123456", fixture: "ccpm/plugin-update.golden"},
		{args: "ccpm plugin update invalid-id --name updated-name --environment env-123456", fixture: "ccpm/plugin-update-not-found.golden", exitCode: 1},

		// Plugin delete tests - using exact IDs from test-server
		{args: "ccpm plugin delete ccp-123456 --environment env-123456 --force", fixture: "ccpm/plugin-delete.golden"},
		{args: "ccpm plugin delete ccp-123456 --environment env-123456", input: "y\n", fixture: "ccpm/plugin-delete-prompt.golden"},
		{args: "ccpm plugin delete ccp-123456 --environment env-123456", input: "n\n", fixture: "ccpm/plugin-delete-cancelled.golden"},
		{args: "ccpm plugin delete invalid-id --environment env-123456 --force", fixture: "ccpm/plugin-delete-not-found.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestCCPMPluginVersion() {
	tests := []CLITest{
		// Version list tests - using exact plugin IDs from test-server
		{args: "ccpm plugin version list --plugin ccp-123456 --environment env-123456", fixture: "ccpm/plugin-version-list.golden"},
		{args: "ccpm plugin version list --plugin ccp-123456 --environment env-123456 -o json", fixture: "ccpm/plugin-version-list-json.golden"},
		{args: "ccpm plugin version list --plugin ccp-123456 --environment env-123456 -o yaml", fixture: "ccpm/plugin-version-list-yaml.golden"},

		// Version create tests - using exact plugin IDs from test-server
		{args: "ccpm plugin version create --plugin ccp-123456 --version 1.0.0 --plugin-file test/fixtures/input/connect/confluentinc-kafka-connect-datagen-0.6.1.zip --connector-classes 'io.confluent.kafka.connect.datagen.DatagenConnector:SOURCE' --environment env-123456", fixture: "ccpm/plugin-version-create.golden"},
		{args: "ccpm plugin version create --plugin ccp-123456 --version 1.0.0 --plugin-file test/fixtures/input/connect/confluentinc-kafka-connect-datagen-0.6.1.zip --connector-classes 'io.confluent.kafka.connect.datagen.DatagenConnector:SOURCE' --sensitive-properties 'password,secret' --documentation-link 'https://docs.confluent.io' --environment env-123456", fixture: "ccpm/plugin-version-create-with-options.golden"},
		{args: "ccpm plugin version create --plugin ccp-123456 --version 1.0.0 --plugin-file test/fixtures/input/connect/invalid-file.txt --connector-classes 'io.confluent.kafka.connect.datagen.DatagenConnector:SOURCE' --environment env-123456", fixture: "ccpm/plugin-version-create-invalid-file.golden", exitCode: 1},

		// Version describe tests - using exact IDs from test-server
		{args: "ccpm plugin version describe ver-123456 --plugin ccp-123456 --environment env-123456", fixture: "ccpm/plugin-version-describe.golden"},
		{args: "ccpm plugin version describe ver-123456 --plugin ccp-123456 --environment env-123456 -o json", fixture: "ccpm/plugin-version-describe-json.golden"},
		{args: "ccpm plugin version describe ver-123456 --plugin ccp-123456 --environment env-123456 -o yaml", fixture: "ccpm/plugin-version-describe-yaml.golden"},
		{args: "ccpm plugin version describe invalid-version --plugin ccp-123456 --environment env-123456", fixture: "ccpm/plugin-version-describe-not-found.golden", exitCode: 1},

		// Version delete tests - using exact IDs from test-server
		{args: "ccpm plugin version delete ver-123456 --plugin ccp-123456 --environment env-123456 --force", fixture: "ccpm/plugin-version-delete.golden"},
		{args: "ccpm plugin version delete ver-123456 --plugin ccp-123456 --environment env-123456", input: "y\n", fixture: "ccpm/plugin-version-delete-prompt.golden"},
		{args: "ccpm plugin version delete ver-123456 --plugin ccp-123456 --environment env-123456", input: "n\n", fixture: "ccpm/plugin-version-delete-cancelled.golden"},
		{args: "ccpm plugin version delete invalid-version --plugin ccp-123456 --environment env-123456 --force", fixture: "ccpm/plugin-version-delete-not-found.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestCCPM_Validation() {
	tests := []CLITest{
		// Required flag validation tests
		{args: "ccpm plugin list", fixture: "ccpm/plugin-list-missing-environment.golden"},
		{args: "ccpm plugin create --name my-plugin --cloud aws", fixture: "ccpm/plugin-create-missing-environment.golden"},
		{args: "ccpm plugin create --name my-plugin --environment env-123456", fixture: "ccpm/plugin-create-missing-cloud.golden", exitCode: 1},
		{args: "ccpm plugin create --cloud aws --environment env-123456", fixture: "ccpm/plugin-create-missing-name.golden", exitCode: 1},
		{args: "ccpm plugin describe ccp-123456", fixture: "ccpm/plugin-describe-missing-environment.golden"},
		{args: "ccpm plugin update ccp-123456", fixture: "ccpm/plugin-update-missing-environment.golden"},
		{args: "ccpm plugin delete ccp-123456 --force", fixture: "ccpm/plugin-delete-missing-environment.golden"},

		// Version command validation tests
		{args: "ccpm plugin version list", fixture: "ccpm/plugin-version-list-missing-plugin.golden", exitCode: 1},
		{args: "ccpm plugin version list --plugin ccp-123456", fixture: "ccpm/plugin-version-list-missing-environment.golden"},
		{args: "ccpm plugin version create --plugin ccp-123456 --version 1.0.0", fixture: "ccpm/plugin-version-create-missing-required.golden", exitCode: 1},
		{args: "ccpm plugin version describe --plugin ccp-123456", fixture: "ccpm/plugin-version-describe-missing-version.golden", exitCode: 1},
		{args: "ccpm plugin version describe ver-123456", fixture: "ccpm/plugin-version-describe-missing-plugin.golden", exitCode: 1},
		{args: "ccpm plugin version delete --plugin ccp-123456", fixture: "ccpm/plugin-version-delete-missing-version.golden", exitCode: 1},
		{args: "ccpm plugin version delete ver-123456", fixture: "ccpm/plugin-version-delete-missing-plugin.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestCCPM_ErrorHandling() {
	tests := []CLITest{
		// Network/API error handling - use different environment IDs for error cases
		{args: "ccpm plugin list --environment env-error", fixture: "ccpm/plugin-list-api-error.golden", exitCode: 1},
		{args: "ccpm plugin create --name my-plugin --cloud aws --environment env-error", fixture: "ccpm/plugin-create-api-error.golden", exitCode: 1},
		{args: "ccpm plugin describe ccp-123456 --environment env-error", fixture: "ccpm/plugin-describe-api-error.golden", exitCode: 1},
		{args: "ccpm plugin update ccp-123456 --name updated-name --environment env-error", fixture: "ccpm/plugin-update-api-error.golden", exitCode: 1},
		{args: "ccpm plugin delete ccp-123456 --environment env-error --force", fixture: "ccpm/plugin-delete-api-error.golden", exitCode: 1},

		// Invalid argument handling
		{args: "ccpm plugin describe", fixture: "ccpm/plugin-describe-missing-id.golden", exitCode: 1},
		{args: "ccpm plugin describe ccp-123456 ccp-789012 --environment env-123456", fixture: "ccpm/plugin-describe-too-many-args.golden", exitCode: 1},
		{args: "ccpm plugin update", fixture: "ccpm/plugin-update-missing-id.golden", exitCode: 1},
		{args: "ccpm plugin update ccp-123456 ccp-789012 --environment env-123456", fixture: "ccpm/plugin-update-too-many-args.golden", exitCode: 1},
		{args: "ccpm plugin delete", fixture: "ccpm/plugin-delete-missing-id.golden", exitCode: 1},
		{args: "ccpm plugin delete ccp-123456 ccp-789012 --environment env-123456", fixture: "ccpm/plugin-delete-too-many-args.golden", exitCode: 1},

		// Version command invalid argument handling
		{args: "ccpm plugin version describe", fixture: "ccpm/plugin-version-describe-missing-id.golden", exitCode: 1},
		{args: "ccpm plugin version describe ver-123456 ver-789012 --plugin ccp-123456 --environment env-123456", fixture: "ccpm/plugin-version-describe-too-many-args.golden", exitCode: 1},
		{args: "ccpm plugin version delete", fixture: "ccpm/plugin-version-delete-missing-id.golden", exitCode: 1},
		{args: "ccpm plugin version delete ver-123456 ver-789012 --plugin ccp-123456 --environment env-123456", fixture: "ccpm/plugin-version-delete-too-many-args.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
