package test

func (s *CLITestSuite) TestRtceRegionList() {
	tests := []CLITest{
		{args: "rtce region list", fixture: "rtce/region/list.golden"},
		{args: "rtce region list --region us-east-2", fixture: "rtce/region/list-region.golden"},
		{args: "rtce region list -o json", fixture: "rtce/region/list-json.golden"},
		{args: "rtce region list -o yaml", fixture: "rtce/region/list-yaml.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
