package test

func (s *CLITestSuite) TestConfluentIAM() {
	tests := []CLITest{
		{Args: "iam role describe --help", Fixture: "iam/confluent-iam-role-describe-help.golden"},
		{Args: "iam role describe DeveloperRead -o json", Fixture: "iam/confluent-iam-role-describe-json.golden"},
		{Args: "iam role describe DeveloperRead -o yaml", Fixture: "iam/confluent-iam-role-describe-yaml.golden"},
		{Args: "iam role describe DeveloperRead", Fixture: "iam/confluent-iam-role-describe.golden"},
		{Args: "iam role list --help", Fixture: "iam/confluent-iam-role-list-help.golden"},
		{Args: "iam role list -o json", Fixture: "iam/confluent-iam-role-list-json.golden"},
		{Args: "iam role list -o yaml", Fixture: "iam/confluent-iam-role-list-yaml.golden"},
		{Args: "iam role list", Fixture: "iam/confluent-iam-role-list.golden"},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestCcloudIAM() {
	tests := []CLITest{
		{Args: "iam role describe CloudClusterAdmin -o json", Fixture: "iam/ccloud-iam-role-describe-json.golden"},
		{Args: "iam role describe CloudClusterAdmin -o yaml", Fixture: "iam/ccloud-iam-role-describe-yaml.golden"},
		{Args: "iam role describe CloudClusterAdmin", Fixture: "iam/ccloud-iam-role-describe.golden"},
		{Args: "iam role describe InvalidRole", Fixture: "iam/ccloud-iam-role-describe-invalid-role.golden", WantErrCode: 1},
		{Args: "iam role list -o json", Fixture: "iam/ccloud-iam-role-list-json.golden"},
		{Args: "iam role list -o yaml", Fixture: "iam/ccloud-iam-role-list-yaml.golden"},
		{Args: "iam role list", Fixture: "iam/ccloud-iam-role-list.golden"},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunCcloudTest(tt)
	}
}
