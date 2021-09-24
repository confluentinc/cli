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

func (s *CLITestSuite) TestIAMServiceAccount() {
	tests := []CLITest{
		{args: "iam service-account create human-service --description human-output", fixture: "service-account/service-account-create.golden"},
		{args: "iam service-account create json-service --description json-output -o json", fixture: "service-account/service-account-create-json.golden"},
		{args: "iam service-account create yaml-service --description yaml-output -o yaml", fixture: "service-account/service-account-create-yaml.golden"},
		{args: "iam service-account delete sa-12345", fixture: "service-account/service-account-delete.golden"},
		{args: "iam service-account list -o json", fixture: "service-account/service-account-list-json.golden"},
		{args: "iam service-account list -o yaml", fixture: "service-account/service-account-list-yaml.golden"},
		{args: "iam service-account list", fixture: "service-account/service-account-list.golden"},
		{args: "iam service-account update sa-12345 --description new-description", fixture: "service-account/service-account-update.golden"},
		{args: "iam service-account update sa-12345 --description new-description-2", fixture: "service-account/service-account-update-2.golden"},
		{args: "iam service-account delete sa-12345", fixture: "service-account/service-account-delete.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCcloudTest(tt)
	}
}

func (s *CLITestSuite) TestIAMUserList() {
	tests := []CLITest{
		{
			args:    "iam user list",
			fixture: "iam/user-list.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}

func (s *CLITestSuite) TestIAMUserDescribe() {
	tests := []CLITest{
		{
			args:        "iam user describe u-0",
			wantErrCode: 1,
			fixture:     "iam/user-resource-not-found.golden",
		},
		{
			args:    "iam user describe u-17",
			fixture: "iam/user-describe.golden",
		},
		{
			args:        "iam user describe 0",
			wantErrCode: 1,
			fixture:     "iam/user-bad-resource-id.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}

func (s *CLITestSuite) TestIAMUserDelete() {
	tests := []CLITest{
		{
			args:    "iam user delete u-0",
			fixture: "iam/user-delete.golden",
		},
		{
			args:        "iam user delete 0",
			wantErrCode: 1,
			fixture:     "iam/user-bad-resource-id.golden",
		},
		{
			args:        "iam user delete u-1",
			wantErrCode: 1,
			fixture:     "iam/user-delete-dne.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}

func (s *CLITestSuite) TestIAMUserInvite() {
	tests := []CLITest{
		{
			args:    "iam user invite miles@confluent.io",
			fixture: "iam/user-invite.golden",
		},
		{
			args:        "iam user invite bad-email.com",
			wantErrCode: 1,
			fixture:     "iam/user-bad-email.golden",
		},
		{
			args:        "iam user invite user@exists.com",
			wantErrCode: 1,
			fixture:     "iam/user-invite-user-already-active.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}
