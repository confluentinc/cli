package test

func (s *CLITestSuite) TestOrganization() {
	tests := []CLITest{
		{args: "organization describe", fixture: "organization/describe.golden"},
		{args: "organization describe -o json", fixture: "organization/describe-json.golden"},
		{args: "organization list", fixture: "organization/list.golden"},
		{args: "organization list -o json", fixture: "organization/list-json.golden"},
		{args: "organization update --name default-updated", fixture: "organization/update.golden"},
		{args: "organization update --name default-updated -o json", fixture: "organization/update-json.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
