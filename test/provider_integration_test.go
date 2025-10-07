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

		// Multiple deletion tests
		{
			args:    "provider-integration v2 delete pi-123456 pi-789012 --force",
			fixture: "provider-integration/v2/delete-multiple-success.golden",
		},
		{
			args:    "provider-integration v2 delete pi-123456 pi-789012",
			input:   "y\n",
			fixture: "provider-integration/v2/delete-multiple-success-prompt.golden",
		},
		{
			args:     "provider-integration v2 delete pi-123456 pi-invalid",
			fixture:  "provider-integration/v2/delete-mixed-invalid.golden",
			exitCode: 1,
		},

		// Create command tests (combined create+authorize flow)
		{
			args:    "provider-integration v2 create azure-test --cloud azure --azure-tenant-id 00000000-0000-0000-0000-000000000000",
			fixture: "provider-integration/v2/create-azure.golden",
		},
		{
			args:    "provider-integration v2 create gcp-test --cloud gcp --gcp-service-account test-sa@test-project.iam.gserviceaccount.com",
			fixture: "provider-integration/v2/create-gcp.golden",
		},

		// Output format tests
		{
			args:    "provider-integration v2 describe pi-123456 --output json",
			fixture: "provider-integration/v2/describe-azure-json.golden",
		},
		{
			args:    "provider-integration v2 describe pi-789012 --output yaml",
			fixture: "provider-integration/v2/describe-gcp-yaml.golden",
		},
		{
			args:    "provider-integration v2 list --output json",
			fixture: "provider-integration/v2/list-json.golden",
		},

		// Environment flag tests
		{
			args:    "provider-integration v2 list --environment env-596",
			fixture: "provider-integration/v2/list-env-flag.golden",
		},
		{
			args:    "provider-integration v2 describe pi-123456 --environment env-596",
			fixture: "provider-integration/v2/describe-env-flag.golden",
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
		{
			args:     "provider-integration v2 describe pi-invalid",
			fixture:  "provider-integration/v2/describe-not-exist.golden",
			exitCode: 1,
		},

		// Atomic behavior tests (invalid configs that should trigger cleanup)
		{
			args:     "provider-integration v2 create atomic-test-invalid-gcp --cloud gcp --gcp-service-account invalid-format",
			fixture:  "provider-integration/v2/create-invalid-gcp-atomic.golden",
			exitCode: 1,
		},
		{
			args:     "provider-integration v2 create atomic-test-invalid-azure --cloud azure --azure-tenant-id not-a-valid-uuid",
			fixture:  "provider-integration/v2/create-invalid-azure-atomic.golden",
			exitCode: 1,
		},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
