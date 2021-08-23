package test

func (s *CLITestSuite) TestConfluentIAMAcl() {
	tests := []CLITest{
		{
			args:    "iam acl create --help",
			fixture: "iam-acl/create-help.golden",
		},
		{
			args:    "iam acl delete --help",
			fixture: "iam-acl/delete-help.golden",
		}, {
			args:    "iam acl list --help",
			fixture: "iam-acl/list-help.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runOnPremTest(tt)
	}
}
