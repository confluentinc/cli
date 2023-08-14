package test

func (s *CLITestSuite) TestNetworkDescribe() {
	tests := []CLITest{
		{args: "network describe n-abcde1", fixture: "network/describe.golden"},
		{args: "network describe n-abcde1 --output json", fixture: "network/describe-json.golden"},
		{args: "network describe n-abcde1 --output yaml", fixture: "network/describe-yaml.golden"},
		{args: "network describe", fixture: "network/describe-missing-id.golden", exitCode: 1},
		{args: "network describe n-invalid", fixture: "network/describe-invalid.golden", exitCode: 1},
	}
	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
