package test

func (s *CLITestSuite) TestEnvironment() {
	tests := []CLITest{
		{args: "environment list", fixture: "environment/1.golden", login: "cloud"},
		{args: "environment use env-595", fixture: "environment/2.golden"},
		{args: "environment update env-595 --name new-other-name --governance-package advanced", fixture: "environment/3.golden"},
		{args: "environment list", fixture: "environment/4.golden"},
		{args: "environment list -o json", fixture: "environment/5.golden"},
		{args: "environment list -o yaml", fixture: "environment/6.golden"},
		{args: "environment use env-dne", fixture: "environment/7.golden", exitCode: 1},
		{args: "environment create saucayyy --governance-package essentials", fixture: "environment/8.golden"},
		{args: "environment create saucayyy -o json --governance-package advanced", fixture: "environment/9.golden"},
		{args: "environment create saucayyy -o yaml", fixture: "environment/10.golden"},
	}

	resetConfiguration(s.T(), false)

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestEnvironmentDescribe() {
	tests := []CLITest{
		{args: "environment describe env-123456", fixture: "environment/describe.golden"},
		{args: "environment describe env-123456 -o json", fixture: "environment/describe-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestEnvironmentDelete() {
	tests := []CLITest{
		{args: "environment delete env-595 --force", fixture: "environment/delete/success.golden"},
		{args: "environment delete env-595", input: "default\n", fixture: "environment/delete/success-prompt.golden"},
		{args: "environment delete env-dne", fixture: "environment/delete/fail.golden", exitCode: 1},
		{args: "environment delete env-srUpdate env-dne", fixture: "environment/delete/multiple-fail.golden", exitCode: 1},
		{args: "environment delete env-srUpdate env-595", input: "n\n", fixture: "environment/delete/multiple-refuse.golden"},
		{args: "environment delete env-srUpdate env-595", input: "y\n", fixture: "environment/delete/multiple-success.golden"},
		{args: "environment delete env-595 env-srUpdate env-000 env-111", input: "y\n", fixture: "environment/delete/multierror.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestEnvironmentUse() {
	tests := []CLITest{
		{args: "environment use env-123456", fixture: "environment/use.golden"},
		{args: "environment describe", fixture: "environment/describe-after-use.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestEnvironmentUpdate_PackageDowngrade() {
	tests := []CLITest{
		{args: "environment update env-595 --governance-package essentials", fixture: "environment/update-governance-package-downgrade-fail.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestEnvironment_Autocomplete() {
	test := CLITest{args: `__complete environment describe ""`, login: "cloud", fixture: "environment/describe-autocomplete.golden"}
	s.runIntegrationTest(test)
}
