package test

func (s *CLITestSuite) TestOrganization() {
	tests := []CLITest{
		{args: "organization describe abc-123", fixture: "organization/describe.golden"},
		{args: "organization describe abc-123 -o json", fixture: "organization/describe-json.golden"},
		{args: "organization describe abc-123 -o yaml", fixture: "organization/describe-yaml.golden"},
		{args: "organization describe abc-456", fixture: "organization/describe-not-current.golden"},
		{args: "organization describe abc-789", fixture: "organization/describe-dne.golden", wantErrCode: 1},
		/*{args: "organization list", fixture: "organization/list.golden"},
		{args: "organization list -o json", fixture: "organization/list-json.golden"},
		{args: "organization list -o yaml", fixture: "organization/list-yaml.golden"},*/
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
