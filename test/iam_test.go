package test

func (s *CLITestSuite) TestConfluentIAM() {
	tests := []CLITest{
		{args: "iam role describe --help", fixture: "iam/confluent-iam-role-describe-help.golden"},
		{args: "iam role describe DeveloperRead -o json", fixture: "iam/confluent-iam-role-describe-json.golden"},
		{args: "iam role describe DeveloperRead -o yaml", fixture: "iam/confluent-iam-role-describe-yaml.golden"},
		{args: "iam role describe DeveloperRead", fixture: "iam/confluent-iam-role-describe.golden"},
		{args: "iam role list --help", fixture: "iam/confluent-iam-role-list-help.golden"},
		{args: "iam role list -o json", fixture: "iam/confluent-iam-role-list-json.golden"},
		{args: "iam role list -o yaml", fixture: "iam/confluent-iam-role-list-yaml.golden"},
		{args: "iam role list", fixture: "iam/confluent-iam-role-list.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestCcloudIAM() {
	tests := []CLITest{
		{args: "iam role describe CloudClusterAdmin -o json", fixture: "iam/ccloud-iam-role-describe-json.golden"},
		{args: "iam role describe CloudClusterAdmin -o yaml", fixture: "iam/ccloud-iam-role-describe-yaml.golden"},
		{args: "iam role describe CloudClusterAdmin", fixture: "iam/ccloud-iam-role-describe.golden"},
		{args: "iam role describe InvalidRole", fixture: "iam/ccloud-iam-role-describe-invalid-role.golden", wantErrCode: 1},
		{args: "iam role list -o json", fixture: "iam/ccloud-iam-role-list-json.golden"},
		{args: "iam role list -o yaml", fixture: "iam/ccloud-iam-role-list-yaml.golden"},
		{args: "iam role list", fixture: "iam/ccloud-iam-role-list.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCcloudTest(tt)
	}
}
