package test

func (s *CLITestSuite) TestConfluentIAMAcl() {
	tests := []CLITest{
		{
			Name:    "confluent iam acl create --help",
			Args:    "iam acl create --help",
			Fixture: "iam-acl/confluent-iam-acl-create-help.golden",
		},
		{
			Name:    "confluent iam acl delete --help",
			Args:    "iam acl delete --help",
			Fixture: "iam-acl/confluent-iam-acl-delete-help.golden",
		}, {
			Name:    "confluent iam acl list --help",
			Args:    "iam acl list --help",
			Fixture: "iam-acl/confluent-iam-acl-list-help.golden",
		},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunConfluentTest(tt)
	}
}
