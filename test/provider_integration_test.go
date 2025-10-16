package test

func (s *CLITestSuite) TestProviderIntegration() {
	tests := []CLITest{
		{args: "provider-integration create cdong-test2 --cloud aws --customer-role-arn arn:aws:iam::037803949979:role/tarun-iam-test-role", fixture: "provider-integration/create.golden"},
		{args: "provider-integration describe cspi-42o61", fixture: "provider-integration/describe.golden"},
		{args: "provider-integration list", fixture: "provider-integration/list.golden"},
		{args: "provider-integration list --cloud aws", fixture: "provider-integration/list.golden"},
		{args: "provider-integration delete cspi-42o61 --force", fixture: "provider-integration/delete-success.golden"},
		{args: "provider-integration delete cspi-42o61", input: "y\n", fixture: "provider-integration/delete-success-prompt.golden"},
		{args: "provider-integration delete cspi-invalid", fixture: "provider-integration/delete-not-exist.golden", exitCode: 1},
		{args: "provider-integration delete cspi-42o61 cspi-43q6j --force", fixture: "provider-integration/delete-multiple-success.golden"},
		{args: "provider-integration delete cspi-42o61 cspi-43q6j", input: "y\n", fixture: "provider-integration/delete-multiple-success-prompt.golden"},
		{args: "provider-integration delete cspi-42o61 cspi-43q6j cspi-invalid", input: "y\n", fixture: "provider-integration/delete-not-exist.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
