package test

func (s *CLITestSuite) TestProviderIntegrationV2() {
	tests := []CLITest{
		// Basic command tests
		{
			args:    "provider-integration v2 list",
			fixture: "provider-integration/v2/list.golden",
		},
		{
			args:    "provider-integration v2 describe pi-123456",
			fixture: "provider-integration/v2/describe-azure.golden",
		},
		{
			args:    "provider-integration v2 describe pi-789012",
			fixture: "provider-integration/v2/describe-gcp.golden",
		},
		{
			args:    "provider-integration v2 delete pi-123456 --force",
			fixture: "provider-integration/v2/delete-success.golden",
		},
		{
			args:    "provider-integration v2 delete pi-123456",
			input:   "y\n",
			fixture: "provider-integration/v2/delete-success-prompt.golden",
		},

		// Error cases
		{
			args:     "provider-integration v2 create invalid-test --cloud invalid",
			fixture:  "provider-integration/v2/create-invalid-provider.golden",
			exitCode: 1,
		},
		{
			args:     "provider-integration v2 create missing-azure-config --cloud azure",
			fixture:  "provider-integration/v2/create-missing-config.golden",
			exitCode: 1,
		},
		{
			args:     "provider-integration v2 create missing-gcp-config --cloud gcp",
			fixture:  "provider-integration/v2/create-missing-gcp-config.golden",
			exitCode: 1,
		},
		{
			args:     "provider-integration v2 delete pi-invalid",
			fixture:  "provider-integration/v2/delete-not-exist.golden",
			exitCode: 1,
		},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
