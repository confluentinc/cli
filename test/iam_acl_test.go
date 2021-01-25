package test

func (s *CLITestSuite) TestConfluentIAMAcl() {
	tests := []CLITest{
		{
			name: "confluent iam acl create --help",
			args: "iam acl create --help",
			fixture: "iam-acl/confluent-iam-acl-create-help.golden",
		},
		{
			name: "confluent iam acl delete --help",
			args: "iam acl delete --help",
			fixture: "iam-acl/confluent-iam-acl-delete-help.golden",
		},		{
			name: "confluent iam acl list --help",
			args: "iam acl list --help",
			fixture: "iam-acl/confluent-iam-acl-list-help.golden",
		},

	}

	for _, tt := range tests {
		tt.login = "default"
		s.runConfluentTest(tt)
	}
}
